import Foundation
import CoreServices

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

@main
@MainActor
struct llc {
    static var forceColor: Bool = false

    static var useColor: Bool {
        if forceColor { return true }
        // 检查是否在终端中且支持颜色
        if let term = getenv("TERM") {
            let termStr = String(cString: term)
            return termStr != "dumb" && isatty(fileno(stdout)) != 0
        }
        return false
    }
    static func main() {
        let arguments = CommandLine.arguments

        var showHidden = false
        var editComment: String? = nil
        var path: String? = nil

        var i = 1
        while i < arguments.count {
            let arg = arguments[i]
            if arg == "-a" {
                showHidden = true
            } else if arg == "--color" {
                forceColor = true
            } else if arg == "-e" {
                // -e 文件夹 "备注信息"
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
            } else if arg == "-h" || arg == "--help" {
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

        // 如果是编辑模式
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
        print("  --color     强制启用颜色输出")
        print("  -e 文件夹 \"备注\"  设置 Finder 注释")
        print("  -h, --help  显示帮助信息")
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
        print("  llc --color            # 强制启用颜色")
        print("  llc -e file.txt \"重要文档\"  # 设置文件注释")
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

        // 根据文件类型选择颜色
        let nameColor: String
        let isExecutable = (permissions & 0o111) != 0
        switch fileType {
        case .typeDirectory:
            nameColor = Colors.blue + Colors.bold
        case .typeSymbolicLink:
            nameColor = Colors.cyan
        default:
            if isExecutable {
                nameColor = Colors.green
            } else {
                nameColor = Colors.reset
            }
        }

        let commentColor = Colors.gray

        if useColor {
            let coloredName = "\(nameColor)\(name)\(Colors.reset)"
            let coloredComment = fileComment.isEmpty ? "" : "  \(commentColor)[\(fileComment)]\(Colors.reset)"

            let output = String(format: "%@ %2d %@ %@ %@ %@ %@%@",
                modeString,
                attrs[.referenceCount] as? Int ?? 1,
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

    static func setFinderComment(path: String, comment: String) {
        let fileManager = FileManager.default

        guard fileManager.fileExists(atPath: path) else {
            print("llc: 文件不存在 '\(path)'")
            exit(1)
        }

        // 使用 AppleScript 设置 Finder 注释
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
