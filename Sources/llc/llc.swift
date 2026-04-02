import Foundation
import CoreServices
import Darwin

// 版本信息
let VERSION = "1.4.0"

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
    case atime
    case ctime
}

// 时间格式
enum TimeStyle {
    case `default`
    case iso
    case longIso
    case fullIso
}

// 颜色设置
enum ColorSetting {
    case always
    case auto
    case never
}

// 配置结构体
struct Config {
    var color: ColorSetting = .auto
    var sort: SortBy = .name
    var groupDirectoriesFirst: Bool = false
    var humanReadable: Bool = false
    var showHidden: Bool = false
    var timeStyle: TimeStyle = .default

    static func load(from path: String) -> Config {
        var config = Config()
        let fileManager = FileManager.default

        guard fileManager.fileExists(atPath: path) else {
            return config
        }

        guard let content = try? String(contentsOfFile: path, encoding: .utf8) else {
            return config
        }

        for line in content.components(separatedBy: .newlines) {
            let trimmed = line.trimmingCharacters(in: .whitespaces)

            // 跳过空行和注释
            if trimmed.isEmpty || trimmed.hasPrefix("#") {
                continue
            }

            // 解析 key = value 格式
            let parts = trimmed.components(separatedBy: "=")
            guard parts.count >= 2 else { continue }

            let key = parts[0].trimmingCharacters(in: .whitespaces).lowercased()
            let value = parts[1...].joined(separator: "=").trimmingCharacters(in: .whitespaces)

            switch key {
            case "color":
                switch value.lowercased() {
                case "always": config.color = .always
                case "auto": config.color = .auto
                case "never": config.color = .never
                default: break
                }
            case "sort":
                switch value.lowercased() {
                case "name": config.sort = .name
                case "time": config.sort = .time
                case "size": config.sort = .size
                case "atime": config.sort = .atime
                case "ctime": config.sort = .ctime
                default: break
                }
            case "group-directories-first":
                config.groupDirectoriesFirst = value.lowercased() == "true" || value == "1"
            case "human-readable":
                config.humanReadable = value.lowercased() == "true" || value == "1"
            case "show-hidden":
                config.showHidden = value.lowercased() == "true" || value == "1"
            case "time-style":
                switch value.lowercased() {
                case "default": config.timeStyle = .default
                case "iso": config.timeStyle = .iso
                case "long-iso": config.timeStyle = .longIso
                case "full-iso": config.timeStyle = .fullIso
                default: break
                }
            default:
                break
            }
        }

        return config
    }
}

// 获取 Finder 注释 - 全局函数避免 actor 隔离问题
func getFinderComment(path: String) -> String {
    let nsUrl = URL(fileURLWithPath: path) as NSURL
    guard let metadataItem = MDItemCreateWithURL(nil, nsUrl as CFURL) else {
        return ""
    }
    guard let comment = MDItemCopyAttribute(metadataItem, kMDItemFinderComment) else {
        return ""
    }
    return comment as? String ?? ""
}

@main
struct llc {
    // 使用实例属性而非静态属性
    var forceColor: Bool = false
    var disableColor: Bool = false
    var configColor: ColorSetting = .auto
    var timeStyle: TimeStyle = .default

    var useColor: Bool {
        if getenv("NO_COLOR") != nil { return false }
        if disableColor { return false }
        if forceColor { return true }
        switch configColor {
        case .always:
            return true
        case .never:
            return false
        case .auto:
            if let term = getenv("TERM") {
                let termStr = String(cString: term)
                return termStr != "dumb" && isatty(fileno(stdout)) != 0
            }
            return false
        }
    }

    static func main() {
        var instance = llc()
        instance.run()
    }

    mutating func run() {
        let arguments = CommandLine.arguments

        // 加载配置文件
        let homeDir = FileManager.default.homeDirectoryForCurrentUser.path
        let configPath = (homeDir as NSString).appendingPathComponent(".llcrc")
        let config = Config.load(from: configPath)

        // 应用配置文件默认值
        var showHidden = config.showHidden
        var showAlmostAll = false
        var showInode = false
        var listDirectoryItself = false
        var humanReadable = config.humanReadable
        var classify = false
        var recursive = false
        var sortBy: SortBy = config.sort
        var reverseSort = false
        var groupDirectoriesFirst = config.groupDirectoriesFirst
        var singleColumn = false
        var editComment: String? = nil
        var path: String? = nil

        // 应用颜色配置
        configColor = config.color

        // 应用时间格式配置
        timeStyle = config.timeStyle

        var i = 1
        while i < arguments.count {
            let arg = arguments[i]
            if arg.hasPrefix("-") && arg.count > 1 && !arg.hasPrefix("--") {
                // 解析组合选项如 -li
                let flags = arg.dropFirst()
                for flag in flags {
                    switch flag {
                    case "a": showHidden = true
                    case "A": showAlmostAll = true
                    case "i": showInode = true
                    case "d": listDirectoryItself = true
                    case "h": humanReadable = true
                    case "F": classify = true
                    case "t": sortBy = .time
                    case "S": sortBy = .size
                    case "u": sortBy = .atime
                    case "c": sortBy = .ctime
                    case "r": reverseSort = true
                    case "R": recursive = true
                    case "l": break // -l 是默认行为，不需要处理
                    case "1": singleColumn = true
                    default:
                        print("llc: 无效选项 -- '\(flag)'")
                        exit(1)
                    }
                }
            } else if arg == "--human-readable" {
                humanReadable = true
            } else if arg == "--color" {
                forceColor = true
            } else if arg == "--no-color" {
                disableColor = true
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
            } else if arg == "--group-directories-first" {
                groupDirectoriesFirst = true
            } else if arg.hasPrefix("--time-style=") {
                let value = String(arg.dropFirst("--time-style=".count))
                switch value.lowercased() {
                case "default": timeStyle = .default
                case "iso": timeStyle = .iso
                case "long-iso": timeStyle = .longIso
                case "full-iso": timeStyle = .fullIso
                default:
                    print("llc: 无效的时间格式 -- '\(value)'")
                    print("有效的格式: default, iso, long-iso, full-iso")
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
            if recursive {
                listDirectoryRecursive(path: expandedPath, showHidden: showHidden, showAlmostAll: showAlmostAll, humanReadable: humanReadable, sortBy: sortBy, reverseSort: reverseSort, showInode: showInode, classify: classify, singleColumn: singleColumn)
            } else if listDirectoryItself {
                if singleColumn {
                    print(getFileNameWithColor(path: expandedPath))
                } else {
                    listFile(path: expandedPath, humanReadable: humanReadable, showInode: showInode, classify: classify)
                }
            } else if singleColumn {
                listDirectorySingleColumn(path: expandedPath, showHidden: showHidden, showAlmostAll: showAlmostAll, sortBy: sortBy, reverseSort: reverseSort, groupDirectoriesFirst: groupDirectoriesFirst, classify: classify)
            } else {
                listDirectory(path: expandedPath, showHidden: showHidden, showAlmostAll: showAlmostAll, humanReadable: humanReadable, sortBy: sortBy, reverseSort: reverseSort, groupDirectoriesFirst: groupDirectoriesFirst, showInode: showInode, classify: classify)
            }
        } else {
            if singleColumn {
                print(getFileNameWithColor(path: expandedPath))
            } else {
                listFile(path: expandedPath, humanReadable: humanReadable, showInode: showInode, classify: classify)
            }
        }
    }

    func printVersion() {
        print("llc version \(VERSION)")
        print("macOS enhanced ls command with Finder comments")
    }

    func printHelp() {
        print("用法: llc [选项] [路径]")
        print("")
        print("选项:")
        print("  -a              显示所有文件，包括隐藏文件（包括 . 和 ..）")
        print("  -A              显示所有文件，包括隐藏文件（不包括 . 和 ..）")
        print("  -1              单列输出（每行一个文件名）")
        print("  -i              显示文件的 inode 号")
        print("  -d              列出目录本身，而非其内容")
        print("  -h, --human-readable  以人类可读格式显示文件大小 (KB, MB, GB)")
        print("  -F              在文件名后添加类型指示符 (*/=@|)")
        print("  -t              按修改时间排序（最新的在前）")
        print("  -S              按文件大小排序（最大的在前）")
        print("  -r              反向排序")
        print("  -R              递归列出子目录")
        print("  --group-directories-first  目录排在文件前面")
        print("  --time-style=STYLE  时间显示格式: default, iso, long-iso, full-iso")
        print("  --color         强制启用颜色输出")
        print("  --no-color      禁用颜色输出")
        print("  -e 文件 \"备注\"  设置 Finder 注释")
        print("  --help          显示帮助信息")
        print("  --version       显示版本信息")
        print("")
        print("配置文件 (~/.llcrc):")
        print("  color = always|auto|never")
        print("  sort = name|time|size|atime|ctime")
        print("  group-directories-first = true|false")
        print("  human-readable = true|false")
        print("  show-hidden = true|false")
        print("  time-style = default|iso|long-iso|full-iso")
        print("")
        print("环境变量:")
        print("  NO_COLOR=1      禁用颜色输出")
        print("")
        print("颜色说明:")
        print("  蓝色粗体 = 目录")
        print("  绿色     = 可执行文件")
        print("  青色     = 符号链接")
        print("  灰色     = 注释")
        print("")
        print("类型指示符 (使用 -F):")
        print("  /  = 目录")
        print("  *  = 可执行文件")
        print("  @  = 符号链接")
        print("  =  = 套接字")
        print("  |  = FIFO")
        print("")
        print("示例:")
        print("  llc                    # 列出当前目录")
        print("  llc -a                 # 列出所有文件")
        print("  llc -lh                # 人类可读大小")
        print("  llc -lt                # 按时间排序")
        print("  llc -li                # 显示 inode 号")
        print("  llc -ld /tmp           # 列出目录本身")
        print("  llc -F                 # 显示类型指示符")
        print("  llc -R                 # 递归列出子目录")
        print("  llc -e file.txt \"备注\" # 设置注释")
    }

    func listDirectory(path: String, showHidden: Bool, showAlmostAll: Bool, humanReadable: Bool, sortBy: SortBy, reverseSort: Bool, groupDirectoriesFirst: Bool = false, showInode: Bool, classify: Bool = false) {
        let fileManager = FileManager.default
        do {
            var contents = try fileManager.contentsOfDirectory(atPath: path)

            if showAlmostAll {
                // -A: 显示所有文件，但排除 . 和 ..
                contents = contents.filter { !$0.hasPrefix(".") } + contents.filter { $0.hasPrefix(".") && $0 != "." && $0 != ".." }
            } else if !showHidden {
                // 默认: 不显示隐藏文件
                contents = contents.filter { !$0.hasPrefix(".") }
            } else {
                // -a: 显示所有文件，包括 . 和 ..
                contents.insert(".", at: 0)
                contents.insert("..", at: 1)
            }

            var fileInfos: [(name: String, path: String, attrs: [FileAttributeKey: Any], isDir: Bool)] = []
            for item in contents {
                let fullPath = (path as NSString).appendingPathComponent(item)
                if let attrs = try? fileManager.attributesOfItem(atPath: fullPath) {
                    let fileType = attrs[.type] as? FileAttributeType ?? .typeRegular
                    let isDir = fileType == .typeDirectory
                    fileInfos.append((name: item, path: fullPath, attrs: attrs, isDir: isDir))
                }
            }

            fileInfos.sort { (a: (name: String, path: String, attrs: [FileAttributeKey: Any], isDir: Bool), b: (name: String, path: String, attrs: [FileAttributeKey: Any], isDir: Bool)) -> Bool in
                // 如果启用目录优先，先按目录/文件排序
                if groupDirectoriesFirst && a.isDir != b.isDir {
                    return a.isDir && !b.isDir
                }

                switch sortBy {
                case .name:
                    return a.name.localizedStandardCompare(b.name) == .orderedAscending
                case .time:
                    let time0 = a.attrs[.modificationDate] as? Date ?? Date.distantPast
                    let time1 = b.attrs[.modificationDate] as? Date ?? Date.distantPast
                    return time0 > time1
                case .size:
                    let size0 = a.attrs[.size] as? Int64 ?? 0
                    let size1 = b.attrs[.size] as? Int64 ?? 0
                    return size0 > size1
                case .atime:
                    // atime (访问时间) 使用 modificationDate 作为替代
                    let time0 = a.attrs[.modificationDate] as? Date ?? Date.distantPast
                    let time1 = b.attrs[.modificationDate] as? Date ?? Date.distantPast
                    return time0 > time1
                case .ctime:
                    // ctime 在 macOS 上没有直接支持，使用 modificationDate 作为替代
                    let time0 = a.attrs[.modificationDate] as? Date ?? Date.distantPast
                    let time1 = b.attrs[.modificationDate] as? Date ?? Date.distantPast
                    return time0 > time1
                }
            }

            if reverseSort {
                fileInfos.reverse()
            }

            let fullPaths = fileInfos.map { $0.path }
            let comments = parallelGetComments(paths: fullPaths)

            for (index, info) in fileInfos.enumerated() {
                listFile(path: info.path, attrs: info.attrs, comment: comments[index], humanReadable: humanReadable, showInode: showInode, classify: classify)
            }
        } catch {
            print("llc: cannot open directory '\(path)': \(error.localizedDescription)")
            exit(1)
        }
    }

    func parallelGetComments(paths: [String]) -> [String] {
        final class ResultBox: @unchecked Sendable {
            var results: [String]
            let lock = NSLock()

            init(count: Int) {
                self.results = Array(repeating: "", count: count)
            }

            func setComment(_ comment: String, at index: Int) {
                lock.lock()
                results[index] = comment
                lock.unlock()
            }
        }

        let resultBox = ResultBox(count: paths.count)
        let group = DispatchGroup()

        for (index, path) in paths.enumerated() {
            group.enter()
            DispatchQueue.global().async {
                autoreleasepool {
                    let comment = getFinderComment(path: path)
                    resultBox.setComment(comment, at: index)
                }
                group.leave()
            }
        }

        group.wait()
        return resultBox.results
    }

    func getTypeIndicator(fileType: FileAttributeType, permissions: Int, path: String) -> String {
        switch fileType {
        case .typeDirectory:
            return "/"
        case .typeSymbolicLink:
            return "@"
        default:
            // 检查是否为套接字或FIFO
            var statBuf = stat()
            if stat(path, &statBuf) == 0 {
                let mode = statBuf.st_mode
                if (mode & S_IFMT) == S_IFSOCK {
                    return "="
                }
                if (mode & S_IFMT) == S_IFIFO {
                    return "|"
                }
            }
            // 检查是否可执行
            if (permissions & 0o111) != 0 {
                return "*"
            }
            return ""
        }
    }

    func listFile(path: String, attrs: [FileAttributeKey: Any]? = nil, comment: String? = nil, humanReadable: Bool = false, showInode: Bool = false, classify: Bool = false) {
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

        // 使用 stat 获取 inode
        var inode: UInt64 = 0
        if showInode {
            var statBuf = stat()
            let result = stat(path, &statBuf)
            if result == 0 {
                inode = UInt64(statBuf.st_ino)
            }
        }

        var output = ""

        if showInode {
            output += String(format: "%10llu ", inode)
        }

        let modeString = modeToString(type: fileType, mode: permissions)
        let sizeString = humanReadable ? formatSizeHumanReadable(size) : formatSize(size)
        let dateString = formatDate(modDate, style: timeStyle)
        let name = (path as NSString).lastPathComponent
        let fileComment = comment ?? getFinderComment(path: path)

        // 获取类型指示符
        let typeIndicator = classify ? getTypeIndicator(fileType: fileType, permissions: permissions, path: path) : ""

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
            let coloredName = "\(nameColor)\(name)\(typeIndicator)\(Colors.reset)"
            let coloredComment = fileComment.isEmpty ? "" : "  \(commentColor)[\(fileComment)]\(Colors.reset)"

            output += String(format: "%@ %2d %@ %@ %@ %@ %@%@",
                modeString,
                fileAttrs[.referenceCount] as? Int ?? 1,
                owner.padding(toLength: 8, withPad: " ", startingAt: 0),
                group.padding(toLength: 8, withPad: " ", startingAt: 0),
                sizeString,
                dateString,
                coloredName,
                coloredComment
            )
        } else {
            output += String(format: "%@ %2d %@ %@ %@ %@ %@%@",
                modeString,
                fileAttrs[.referenceCount] as? Int ?? 1,
                owner.padding(toLength: 8, withPad: " ", startingAt: 0),
                group.padding(toLength: 8, withPad: " ", startingAt: 0),
                sizeString,
                dateString,
                name,
                typeIndicator
            )

            if !fileComment.isEmpty {
                output += "  [\(fileComment)]"
            }
        }

        print(output)
    }

    func modeToString(type: FileAttributeType, mode: Int) -> String {
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

    func formatSize(_ size: Int64) -> String {
        return String(format: "%8lld", size)
    }

    func formatSizeHumanReadable(_ size: Int64) -> String {
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

    func formatDate(_ date: Date, style: TimeStyle = .default) -> String {
        let formatter = DateFormatter()
        let calendar = Calendar.current

        switch style {
        case .default:
            if calendar.isDate(date, equalTo: Date(), toGranularity: .year) {
                formatter.dateFormat = "MMM dd HH:mm"
            } else {
                formatter.dateFormat = "MMM dd  yyyy"
            }
            formatter.locale = Locale(identifier: "en_US")
        case .iso:
            formatter.dateFormat = "yyyy-MM-dd"
            formatter.locale = Locale(identifier: "en_US")
        case .longIso:
            formatter.dateFormat = "yyyy-MM-dd HH:mm"
            formatter.locale = Locale(identifier: "en_US")
        case .fullIso:
            formatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
            formatter.timeZone = TimeZone.current
            formatter.locale = Locale(identifier: "en_US")
            let dateStr = formatter.string(from: date)
            // 添加时区偏移 (如 +0800)
            let tzFormatter = DateFormatter()
            tzFormatter.dateFormat = "Z"
            tzFormatter.timeZone = TimeZone.current
            let tzStr = tzFormatter.string(from: date)
            return "\(dateStr) \(tzStr)"
        }

        return formatter.string(from: date)
    }

    func listDirectoryRecursive(path: String, showHidden: Bool, showAlmostAll: Bool, humanReadable: Bool, sortBy: SortBy, reverseSort: Bool, showInode: Bool, classify: Bool, singleColumn: Bool = false, visitedPaths: Set<String> = [], depth: Int = 0) {
        // 防止循环引用和过深层级
        let canonicalPath = (path as NSString).standardizingPath
        if visitedPaths.contains(canonicalPath) || depth > 10 {
            return
        }
        var newVisitedPaths = visitedPaths
        newVisitedPaths.insert(canonicalPath)

        // 打印当前目录路径
        if depth > 0 {
            print("")
        }
        print("\(path):")

        // 先列出当前目录内容
        if singleColumn {
            listDirectorySingleColumn(path: path, showHidden: showHidden, showAlmostAll: showAlmostAll, sortBy: sortBy, reverseSort: reverseSort, groupDirectoriesFirst: false, classify: classify)
        } else {
            listDirectory(path: path, showHidden: showHidden, showAlmostAll: showAlmostAll, humanReadable: humanReadable, sortBy: sortBy, reverseSort: reverseSort, showInode: showInode, classify: classify)
        }

        // 获取子目录列表
        let fileManager = FileManager.default
        do {
            let contents = try fileManager.contentsOfDirectory(atPath: path)
            for item in contents {
                if item.hasPrefix(".") && !showHidden && !showAlmostAll {
                    continue
                }
                let fullPath = (path as NSString).appendingPathComponent(item)
                var isDir: ObjCBool = false
                if fileManager.fileExists(atPath: fullPath, isDirectory: &isDir) {
                    if isDir.boolValue {
                        // 递归处理子目录
                        listDirectoryRecursive(path: fullPath, showHidden: showHidden, showAlmostAll: showAlmostAll, humanReadable: humanReadable, sortBy: sortBy, reverseSort: reverseSort, showInode: showInode, classify: classify, singleColumn: singleColumn, visitedPaths: newVisitedPaths, depth: depth + 1)
                    }
                }
            }
        } catch {
            // 忽略无法访问的目录
        }
    }

    func getFileNameWithColor(path: String, classify: Bool = false) -> String {
        let fileManager = FileManager.default
        let name = (path as NSString).lastPathComponent

        guard let attrs = try? fileManager.attributesOfItem(atPath: path) else {
            return name
        }

        let fileType = attrs[.type] as? FileAttributeType ?? .typeRegular
        let permissions = attrs[.posixPermissions] as? Int ?? 0
        let isExecutable = (permissions & 0o111) != 0

        // 获取类型指示符
        let typeIndicator = classify ? getTypeIndicator(fileType: fileType, permissions: permissions, path: path) : ""

        if !useColor {
            return name + typeIndicator
        }

        let nameColor: String
        switch fileType {
        case .typeDirectory:
            nameColor = Colors.blue + Colors.bold
        case .typeSymbolicLink:
            nameColor = Colors.cyan
        default:
            nameColor = isExecutable ? Colors.green : Colors.reset
        }

        return "\(nameColor)\(name)\(typeIndicator)\(Colors.reset)"
    }

    func listDirectorySingleColumn(path: String, showHidden: Bool, showAlmostAll: Bool, sortBy: SortBy, reverseSort: Bool, groupDirectoriesFirst: Bool = false, classify: Bool = false) {
        let fileManager = FileManager.default
        do {
            var contents = try fileManager.contentsOfDirectory(atPath: path)

            if showAlmostAll {
                // -A: 显示所有文件，但排除 . 和 ..
                contents = contents.filter { !$0.hasPrefix(".") } + contents.filter { $0.hasPrefix(".") && $0 != "." && $0 != ".." }
            } else if !showHidden {
                // 默认: 不显示隐藏文件
                contents = contents.filter { !$0.hasPrefix(".") }
            } else {
                // -a: 显示所有文件，包括 . 和 ..
                contents.insert(".", at: 0)
                contents.insert("..", at: 1)
            }

            var fileInfos: [(name: String, path: String, attrs: [FileAttributeKey: Any], isDir: Bool)] = []
            for item in contents {
                let fullPath = (path as NSString).appendingPathComponent(item)
                if let attrs = try? fileManager.attributesOfItem(atPath: fullPath) {
                    let fileType = attrs[.type] as? FileAttributeType ?? .typeRegular
                    let isDir = fileType == .typeDirectory
                    fileInfos.append((name: item, path: fullPath, attrs: attrs, isDir: isDir))
                }
            }

            fileInfos.sort { (a: (name: String, path: String, attrs: [FileAttributeKey: Any], isDir: Bool), b: (name: String, path: String, attrs: [FileAttributeKey: Any], isDir: Bool)) -> Bool in
                // 如果启用目录优先，先按目录/文件排序
                if groupDirectoriesFirst && a.isDir != b.isDir {
                    return a.isDir && !b.isDir
                }

                switch sortBy {
                case .name:
                    return a.name.localizedStandardCompare(b.name) == .orderedAscending
                case .time:
                    let time0 = a.attrs[.modificationDate] as? Date ?? Date.distantPast
                    let time1 = b.attrs[.modificationDate] as? Date ?? Date.distantPast
                    return time0 > time1
                case .size:
                    let size0 = a.attrs[.size] as? Int64 ?? 0
                    let size1 = b.attrs[.size] as? Int64 ?? 0
                    return size0 > size1
                case .atime:
                    let time0 = a.attrs[.modificationDate] as? Date ?? Date.distantPast
                    let time1 = b.attrs[.modificationDate] as? Date ?? Date.distantPast
                    return time0 > time1
                case .ctime:
                    let time0 = a.attrs[.modificationDate] as? Date ?? Date.distantPast
                    let time1 = b.attrs[.modificationDate] as? Date ?? Date.distantPast
                    return time0 > time1
                }
            }

            if reverseSort {
                fileInfos.reverse()
            }

            for info in fileInfos {
                print(getFileNameWithColor(path: info.path, classify: classify))
            }
        } catch {
            print("llc: cannot open directory '\(path)': \(error.localizedDescription)")
            exit(1)
        }
    }

    func setFinderComment(path: String, comment: String) {
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
