import Foundation
import CoreServices

// 版本信息
let VERSION = "1.1.0"

// ANSI 颜色代码
struct Colors {
    static let reset = "\u{001B}[0m"
    static let bold = "\u{001B}[1m"
    static let black = "\u{001B}[30m"
    static let red = "\u{001B}[31m"
    static let green = "\u{001B}[32m"
    static let yellow = "\u{001B}[33m"
    static let blue = "\u{001B}[34m"
    static let magenta = "\u{001B}[35m"
    static let cyan = "\u{001B}[36m"
    static let white = "\u{001B}[37m"
    static let gray = "\u{001B}[90m"
}

// 排序方式
enum SortBy {
    case name
    case time
    case size
}

@main
@MainActor
struct llc {
    static var forceColor: Bool = false

    static var useColor: Bool {
        if forceColor { return true }
        if let term = getenv("TERM") {
            let termStr = String(cString: term)
            return termStr != "dumb" && isatty(fileno(stdout)) != 0
        }
        return false
    }

    static func main() {
        let arguments = CommandLine.arguments

        var showHidden = false
        var humanReadable = false
        var sortBy: SortBy = .name
        var reverseSort = false
        var editComment: String? = nil
        var path: String? = nil

        var i = 1
        while i < arguments.count {
            let arg = arguments[i]
            if arg == "-a" {
                showHidden = true
            } else if arg == "-h" || arg == "--human-readable" {
                humanReadable = true
            } else if arg == "-t" {
                sortBy = .time
            } else if arg == "-S" {
                sortBy = .size
            } else if arg == "-r" {
                reverseSort = true
            } else if arg == "--color" {
                forceColor = true
            } else if arg == "--version" {
                printVersion()
                exit(0)
            } else if arg == "-e" {
                i += 1
                if i < arguments.count {
                    path = arguments[i]
                    i += 1
                    if i < arguments.count {
                        editComment = arguments[i]
                    } else {
                        print("llc: -e 需要指定备注内容")
                        exit(1)
                    }
                } else {
                    print("llc: -e 需要指定文件夹路径")
                    exit(1)
                }
            } else if arg == "--help" {
                printHelp()
                exit(0)
            } else if !arg.hasPrefix("-") {
                path = arg
            }
            i += 1
        }

        let targetPath = path ?? "."
        let fileManager = FileManager.default
        let expandedPath = (targetPath as NSString).expandingTildeInPath

        if let comment = editComment {
            setFinderComment(path: expandedPath, comment: comment)
            exit(0)
        }

        var isDirectory: ObjCBool = false
        guard fileManager.fileExists(atPath: expandedPath, isDirectory: &isDirectory) else {
            print("llc: cannot access '\(targetPath)': No such file or directory")
            exit(1)
        }

        if isDirectory.boolValue {
            listDirectory(path: expandedPath, showHidden: showHidden, humanReadable: humanReadable, sortBy: sortBy, reverseSort: reverseSort)
        } else {
            listFile(path: expandedPath, humanReadable: humanReadable)
        }
    }

    static func printVersion() {
        print("llc version \(VERSION)")
        print("macOS enhanced ls command with Finder comments")
    }

    static func printHelp() {
        print("用法: llc [选项] [路径]")
        print("")
        print("选项:")
        print("  -a              显示所有文件，包括隐藏文件")
        print("  -h, --human-readable  以人类可读格式显示文件大小 (KB, MB, GB)")
        print("  -t              按修改时间排序（最新的在前）")
        print("  -S              按文件大小排序（最大的在前）")
        print("  -r              反向排序")
        print("  --color         强制启用颜色输出")
        print("  -e 文件 \"备注\"  设置 Finder 注释")
        print("  --help          显示帮助信息")
        print("  --version       显示版本信息")
        print("")
        print("颜色说明:")
        print("  蓝色粗体 = 目录")
        print("  绿色     = 可执行文件")
        print("  青色     = 符号链接")
        print("  灰色     = 注释")
        print("")
        print("示例:")
        print("  llc                    # 列出当前目录")
        print("  llc -a                 # 列出所有文件")
        print("  llc -lh                # 人类可读大小")
        print("  llc -lt                # 按时间排序")
        print("  llc -lS                # 按大小排序")
        print("  llc -ltr               # 按时间反向排序")
        print("  llc -e file.txt \"备注\" # 设置注释")
    }

    static func listDirectory(path: String, showHidden: Bool, humanReadable: Bool, sortBy: SortBy, reverseSort: Bool) {
        let fileManager = FileManager.default
        do {
            var contents = try fileManager.contentsOfDirectory(atPath: path)

            if !showHidden {
                contents = contents.filter { !$0.hasPrefix(".") }
            } else {
                contents.insert(".", at: 0)
                contents.insert("..", at: 1)
            }

            // 获取文件信息用于排序
            var fileInfos: [(name: String, path: String, attrs: [FileAttributeKey: Any])] = []
            for item in contents {
                let fullPath = (path as NSString).appendingPathComponent(item)
                if let attrs = try? fileManager.attributesOfItem(atPath: fullPath) {
                    fileInfos.append((name: item, path: fullPath, attrs: attrs))
                }
            }

            // 排序
            fileInfos.sort {
                switch sortBy {
                case .name:
                    return $0.name.localizedStandardCompare($1.name) == .orderedAscending
                case .time:
                    let time0 = $0.attrs[.modificationDate] as? Date ?? Date.distantPast
                    let time1 = $1.attrs[.modificationDate] as? Date ?? Date.distantPast
                    return time0 > time1
                case .size:
                    let size0 = $0.attrs[.size] as? Int64 ?? 0
                    let size1 = $1.attrs[.size] as? Int64 ?? 0
                    return size0 > size1
                }
            }

            if reverseSort {
                fileInfos.reverse()
            }

            let fullPaths = fileInfos.map { $0.path }
            let comments = parallelGetComments(paths: fullPaths)

            for (index, info) in fileInfos.enumerated() {
                listFile(path: info.path, attrs: info.attrs, comment: comments[index], humanReadable: humanReadable)
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

    static func listFile(path: String, attrs: [FileAttributeKey: Any]? = nil, comment: String? = nil, humanReadable: Bool = false) {
        let fileManager = FileManager.default

        let fileAttrs: [FileAttributeKey: Any]
        if let attrs = attrs {
            fileAttrs = attrs
        } else {
            guard let attrs = try? fileManager.attributesOfItem(atPath: path) else {
                return
            }
            fileAttrs = attrs
        }

        let fileType = fileAttrs[.type] as? FileAttributeType ?? .typeRegular
        let permissions = fileAttrs[.posixPermissions] as? Int ?? 0
        let owner = fileAttrs[.ownerAccountName] as? String ?? "unknown"
        let group = fileAttrs[.groupOwnerAccountName] as? String ?? "unknown"
        let size = fileAttrs[.size] as? Int64 ?? 0
        let modDate = fileAttrs[.modificationDate] as? Date ?? Date()

        let modeString = modeToString(type: fileType, mode: permissions) + " "
        let sizeString = humanReadable ? formatSizeHumanReadable(size) : formatSize(size)
        let dateString = formatDate(modDate)
        let name = (path as NSString).lastPathComponent
        let fileComment = comment ?? getFinderComment(path: path)

        let nameColor: String
        let isExecutable = (permissions & 0o111) != 0
        switch fileType {
        case .typeDirectory:
            nameColor = Colors.blue + Colors.bold
        case .typeSymbolicLink:
            nameColor = Colors.cyan
        default:
            nameColor = isExecutable ? Colors.green : Colors.reset
        }

        let commentColor = Colors.gray

        if useColor {
            let coloredName = "\(nameColor)\(name)\(Colors.reset)"
            let coloredComment = fileComment.isEmpty ? "" : "  \(commentColor)[\(fileComment)]\(Colors.reset)"

            let output = String(format: "%@ %2d %@ %@ %@ %@ %@%@",
                modeString,
                fileAttrs[.referenceCount] as? Int ?? 1,
                owner.padding(toLength: 8, withPad: " ", startingAt: 0),
                group.padding(toLength: 8, withPad: " ", startingAt: 0),
                sizeString,
                dateString,
                coloredName,
                coloredComment
            )
            print(output)
        } else {
            var output = String(format: "%@ %2d %@ %@ %@ %@ %@",
                modeString,
                fileAttrs[.referenceCount] as? Int ?? 1,
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

    static func formatSizeHumanReadable(_ size: Int64) -> String {
        let units = ["B", "K", "M", "G", "T", "P"]
        var value = Double(size)
        var unitIndex = 0

        while value >= 1024 && unitIndex < units.count - 1 {
            value /= 1024
            unitIndex += 1
        }

        if unitIndex == 0 {
            return String(format: "%8lldB", size)
        } else {
            return String(format: "%7.1f%@", value, units[unitIndex])
        }
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

    static func setFinderComment(path: String, comment: String) {
        let fileManager = FileManager.default

        guard fileManager.fileExists(atPath: path) else {
            print("llc: 文件不存在 '\(path)'")
            exit(1)
        }

        let absolutePath = (path as NSString).standardizingPath
        let escapedPath = absolutePath.replacingOccurrences(of: "\"", with: "\\\"")
        let escapedComment = comment.replacingOccurrences(of: "\"", with: "\\\"")

        let appleScript = """
        tell application "Finder"
            set theFile to POSIX file "\(escapedPath)" as alias
            set comment of theFile to "\(escapedComment)"
        end tell
        """

        let process = Process()
        process.launchPath = "/usr/bin/osascript"
        process.arguments = ["-e", appleScript]

        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = pipe

        do {
            try process.run()
            process.waitUntilExit()
        } catch {
            print("llc: 设置注释失败: \(error.localizedDescription)")
            exit(1)
        }

        if process.terminationStatus == 0 {
            print("已设置注释: [\(comment)] -> \(absolutePath)")
        } else {
            let errorData = pipe.fileHandleForReading.readDataToEndOfFile()
            let errorMsg = String(data: errorData, encoding: .utf8) ?? "未知错误"
            print("llc: 设置注释失败: \(errorMsg)")
            exit(1)
        }
    }
}
