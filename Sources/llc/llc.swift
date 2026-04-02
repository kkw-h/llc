import Foundation
import CoreServices

@main
struct llc {
    static func main() {
        let arguments = CommandLine.arguments

        var showHidden = false
        var path: String? = nil

        for arg in arguments.dropFirst() {
            if arg == "-a" {
                showHidden = true
            } else if arg == "-h" || arg == "--help" {
                printHelp()
                exit(0)
            } else if !arg.hasPrefix("-") {
                path = arg
            }
        }

        let targetPath = path ?? "."
        let fileManager = FileManager.default
        let expandedPath = (targetPath as NSString).expandingTildeInPath

        var isDirectory: ObjCBool = false
        guard fileManager.fileExists(atPath: expandedPath, isDirectory: &isDirectory) else {
            print("llc: cannot access '\(targetPath)': No such file or directory")
            exit(1)
        }

        if isDirectory.boolValue {
            listDirectory(path: expandedPath, showHidden: showHidden)
        } else {
            listFile(path: expandedPath)
        }
    }

    static func printHelp() {
        print("用法: llc [选项] [路径]")
        print("")
        print("选项:")
        print("  -a          显示所有文件，包括隐藏文件")
        print("  -h, --help  显示帮助信息")
        print("")
        print("llc 是 ls -l 的增强版本，显示文件列表并在最后显示 Finder 注释")
    }

    static func listDirectory(path: String, showHidden: Bool = false) {
        let fileManager = FileManager.default
        do {
            var contents = try fileManager.contentsOfDirectory(atPath: path)

            if !showHidden {
                contents = contents.filter { !$0.hasPrefix(".") }
            } else {
                contents.insert(".", at: 0)
                contents.insert("..", at: 1)
            }

            let sortedContents = contents.sorted()
            let fullPaths = sortedContents.map { (path as NSString).appendingPathComponent($0) }

            // 并行获取所有文件的 Finder 注释
            let comments = parallelGetComments(paths: fullPaths)

            for (index, _) in sortedContents.enumerated() {
                let fullPath = fullPaths[index]
                let comment = comments[index]
                listFile(path: fullPath, comment: comment)
            }
        } catch {
            print("llc: cannot open directory '\(path)': \(error.localizedDescription)")
            exit(1)
        }
    }

    static func parallelGetComments(paths: [String]) -> [String] {
        var results = Array(repeating: "", count: paths.count)
        let group = DispatchGroup()
        let lock = NSLock()

        for (index, path) in paths.enumerated() {
            group.enter()
            DispatchQueue.global().async {
                let comment = getFinderComment(path: path)
                lock.lock()
                results[index] = comment
                lock.unlock()
                group.leave()
            }
        }

        group.wait()
        return results
    }

    static func listFile(path: String, comment: String? = nil) {
        let fileManager = FileManager.default

        guard let attrs = try? fileManager.attributesOfItem(atPath: path) else {
            return
        }

        let fileType = attrs[.type] as? FileAttributeType ?? .typeRegular
        let permissions = attrs[.posixPermissions] as? Int ?? 0
        let owner = attrs[.ownerAccountName] as? String ?? "unknown"
        let group = attrs[.groupOwnerAccountName] as? String ?? "unknown"
        let size = attrs[.size] as? Int64 ?? 0
        let modDate = attrs[.modificationDate] as? Date ?? Date()

        let modeString = modeToString(type: fileType, mode: permissions) + " "
        let sizeString = formatSize(size)
        let dateString = formatDate(modDate)
        let name = (path as NSString).lastPathComponent
        let fileComment = comment ?? getFinderComment(path: path)

        var output = String(format: "%@ %2d %@ %@ %@ %@ %@",
            modeString,
            attrs[.referenceCount] as? Int ?? 1,
            owner.padding(toLength: 8, withPad: " ", startingAt: 0),
            group.padding(toLength: 8, withPad: " ", startingAt: 0),
            sizeString,
            dateString,
            name
        )

        if !fileComment.isEmpty {
            output += "  [\(fileComment)]"
        }

        print(output)
    }

    static func modeToString(type: FileAttributeType, mode: Int) -> String {
        var result = ""

        switch type {
        case .typeDirectory:
            result = "d"
        case .typeSymbolicLink:
            result = "l"
        default:
            result = "-"
        }

        let permissions = [
            (mode >> 6) & 0o7,
            (mode >> 3) & 0o7,
            mode & 0o7
        ]

        for perm in permissions {
            result += (perm & 0o4) != 0 ? "r" : "-"
            result += (perm & 0o2) != 0 ? "w" : "-"
            result += (perm & 0o1) != 0 ? "x" : "-"
        }

        return result
    }

    static func formatSize(_ size: Int64) -> String {
        return String(format: "%8lld", size)
    }

    static func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        let calendar = Calendar.current

        if calendar.isDate(date, equalTo: Date(), toGranularity: .year) {
            formatter.dateFormat = "MMM dd HH:mm"
        } else {
            formatter.dateFormat = "MMM dd  yyyy"
        }

        formatter.locale = Locale(identifier: "en_US")
        return formatter.string(from: date)
    }

    static func getFinderComment(path: String) -> String {
        let nsUrl = URL(fileURLWithPath: path) as NSURL
        guard let metadataItem = MDItemCreateWithURL(nil, nsUrl as CFURL) else {
            return ""
        }

        guard let comment = MDItemCopyAttribute(metadataItem, kMDItemFinderComment) else {
            return ""
        }

        return comment as? String ?? ""
    }
}
