package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const version = "1.6.0"

// Application constants
const (
	// Concurrency limits
	maxConcurrency = 20

	// Recursion limits
	maxRecursionDepth = 10

	// xattr attribute name
	xattrCommentName = "user.llc.comment"

	// Config file name
	configFileName = ".llcrc"

	// Date/time format strings
	dateFormatDefault     = "Jan 02 15:04"
	dateFormatDefaultYear = "Jan 02  2006"
	dateFormatISO         = "2006-01-02"
	dateFormatLongISO     = "2006-01-02 15:04"
	dateFormatFullISO     = "2006-01-02 15:04:05 -0700"
)

// ANSI color codes
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

// regexCache caches compiled regex patterns for performance
type regexCache struct {
	cache map[string]*regexp.Regexp
	mu    sync.RWMutex
}

var patternCache = &regexCache{
	cache: make(map[string]*regexp.Regexp),
}

func (rc *regexCache) get(pattern string) *regexp.Regexp {
	rc.mu.RLock()
	re, ok := rc.cache[pattern]
	rc.mu.RUnlock()

	if ok {
		return re
	}

	// Compile pattern
	regexPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, ".*"), `\?`, ".") + "$"
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil
	}

	// Store in cache
	rc.mu.Lock()
	rc.cache[pattern] = re
	rc.mu.Unlock()

	return re
}

// Config holds configuration
type Config struct {
	Color          string
	Sort           string
	GroupDirsFirst bool
	HumanReadable  bool
	ShowHidden     bool
	TimeStyle      string
	IgnorePatterns []string
}

// FileInfo holds file information
type FileInfo struct {
	Name    string
	Path    string
	Info    os.FileInfo
	Comment string
	IsDir   bool
}

func main() {
	var (
		showAll         = flag.Bool("a", false, "显示所有文件，包括隐藏文件")
		showAlmostAll   = flag.Bool("A", false, "显示所有文件，不包括 . 和 ..")
		singleColumn    = flag.Bool("1", false, "单列输出")
		showInode       = flag.Bool("i", false, "显示 inode 号")
		listDirSelf     = flag.Bool("d", false, "列出目录本身")
		humanReadable   = flag.Bool("h", false, "人类可读大小")
		classify        = flag.Bool("F", false, "添加类型指示符")
		followSymlinks  = flag.Bool("L", false, "跟随符号链接")
		sortByTime      = flag.Bool("t", false, "按时间排序")
		sortBySize      = flag.Bool("S", false, "按大小排序")
		reverseSort     = flag.Bool("r", false, "反向排序")
		recursive       = flag.Bool("R", false, "递归列出")
		groupDirsFirst  = flag.Bool("group-directories-first", false, "目录排在文件前面")
		colorFlag       = flag.String("color", "", "颜色输出: always, auto, never")
		noColor         = flag.Bool("no-color", false, "禁用颜色")
		timeStyle       = flag.String("time-style", "", "时间格式: default, iso, long-iso, full-iso")
		ignorePattern   = flag.String("ignore", "", "忽略模式")
		editComment     = flag.String("e", "", "设置注释的文件路径")
		editCommentText = flag.String("comment", "", "注释内容（与 -e 配合使用）")
		showVersion     = flag.Bool("version", false, "显示版本")
		showHelp        = flag.Bool("help", false, "显示帮助")
	)
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	if *showHelp {
		printHelp()
		return
	}

	// Load config
	config := loadConfig()

	// Apply config defaults
	if !*humanReadable && config.HumanReadable {
		*humanReadable = true
	}
	if !*showAll && !*showAlmostAll && config.ShowHidden {
		*showAll = true
	}
	if !*groupDirsFirst && config.GroupDirsFirst {
		*groupDirsFirst = true
	}
	if *timeStyle == "" && config.TimeStyle != "" {
		*timeStyle = config.TimeStyle
	}

	// Determine color setting
	useColor := shouldUseColor(*colorFlag, *noColor, config.Color)

	// Handle edit comment mode
	if *editComment != "" {
		args := flag.Args()
		commentText := *editCommentText
		if commentText == "" && len(args) > 0 {
			commentText = args[0]
		}
		if commentText == "" {
			fmt.Fprintf(os.Stderr, "llc: -e 需要指定备注内容\n")
			os.Exit(1)
		}
		setComment(*editComment, commentText)
		return
	}

	// Get target path
	targetPath := "."
	args := flag.Args()
	if len(args) > 0 {
		targetPath = args[0]
	}

	targetPath = expandPath(targetPath)

	info, err := os.Stat(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "llc: cannot access '%s': %v\n", targetPath, err)
		os.Exit(1)
	}

	// Collect ignore patterns
	ignorePatterns := config.IgnorePatterns
	if *ignorePattern != "" {
		ignorePatterns = append(ignorePatterns, *ignorePattern)
	}

	// Determine sort method
	sortBy := "name"
	if *sortByTime {
		sortBy = "time"
	} else if *sortBySize {
		sortBy = "size"
	} else if config.Sort != "" {
		sortBy = config.Sort
	}

	if info.IsDir() {
		if *recursive {
			listRecursive(targetPath, *showAll, *showAlmostAll, *humanReadable, sortBy, *reverseSort, *showInode, *classify, useColor, *timeStyle, ignorePatterns, *groupDirsFirst, *singleColumn, *followSymlinks, 0, make(map[string]bool))
		} else if *listDirSelf {
			if *singleColumn {
				fmt.Println(getColoredName(targetPath, useColor, *classify, *followSymlinks))
			} else {
				listFile(targetPath, nil, "", *humanReadable, *showInode, *classify, useColor, *timeStyle, *followSymlinks)
			}
		} else if *singleColumn {
			listSingleColumn(targetPath, *showAll, *showAlmostAll, sortBy, *reverseSort, *classify, useColor, ignorePatterns, *groupDirsFirst, *followSymlinks)
		} else {
			listDirectory(targetPath, *showAll, *showAlmostAll, *humanReadable, sortBy, *reverseSort, *showInode, *classify, useColor, *timeStyle, ignorePatterns, *groupDirsFirst, *followSymlinks)
		}
	} else {
		if *singleColumn {
			fmt.Println(getColoredName(targetPath, useColor, *classify, *followSymlinks))
		} else {
			// Get comment for single file
			comment := getComment(targetPath)
			listFile(targetPath, info, comment, *humanReadable, *showInode, *classify, useColor, *timeStyle, *followSymlinks)
		}
	}
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return path
}

func shouldUseColor(colorFlag string, noColor bool, configColor string) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if noColor {
		return false
	}
	if colorFlag == "always" {
		return true
	}
	if colorFlag == "never" {
		return false
	}
	if configColor == "always" {
		return true
	}
	if configColor == "never" {
		return false
	}

	// Auto mode - check if stdout is a terminal
	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}
	if term == "" {
		return false
	}

	// Check if stdout is a terminal
	if fi, err := os.Stdout.Stat(); err == nil {
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

func loadConfig() Config {
	config := Config{
		Color:          "auto",
		Sort:           "name",
		TimeStyle:      "default",
		IgnorePatterns: []string{},
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return config
	}

	configPath := filepath.Join(home, configFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "color":
			config.Color = strings.ToLower(value)
		case "sort":
			config.Sort = strings.ToLower(value)
		case "group-directories-first":
			config.GroupDirsFirst = value == "true" || value == "1"
		case "human-readable":
			config.HumanReadable = value == "true" || value == "1"
		case "show-hidden":
			config.ShowHidden = value == "true" || value == "1"
		case "time-style":
			config.TimeStyle = strings.ToLower(value)
		case "ignore":
			config.IgnorePatterns = append(config.IgnorePatterns, value)
		}
	}

	return config
}

func printVersion() {
	fmt.Printf("llc version %s\n", version)
	fmt.Println("Enhanced ls command with file comments support")
	fmt.Println("Platform: Linux/macOS (using xattr/Spotlight)")
}

func printHelp() {
	fmt.Println("用法: llc [选项] [路径]")
	fmt.Println("")
	fmt.Println("选项:")
	fmt.Println("  -a              显示所有文件，包括隐藏文件（包括 . 和 ..）")
	fmt.Println("  -A              显示所有文件，包括隐藏文件（不包括 . 和 ..）")
	fmt.Println("  -1              单列输出（每行一个文件名）")
	fmt.Println("  -i              显示文件的 inode 号")
	fmt.Println("  -d              列出目录本身，而非其内容")
	fmt.Println("  -h              以人类可读格式显示文件大小 (KB, MB, GB)")
	fmt.Println("  -F              在文件名后添加类型指示符 (*/=@|)")
	fmt.Println("  -L              跟随符号链接，显示目标文件信息")
	fmt.Println("  -t              按修改时间排序（最新的在前）")
	fmt.Println("  -S              按文件大小排序（最大的在前）")
	fmt.Println("  -r              反向排序")
	fmt.Println("  -R              递归列出子目录")
	fmt.Println("  --group-directories-first  目录排在文件前面")
	fmt.Println("  --ignore=PATTERN   忽略匹配的文件 (支持 * 和 ? 通配符)")
	fmt.Println("  --time-style=STYLE  时间显示格式: default, iso, long-iso, full-iso")
	fmt.Println("  --color=WHEN    颜色输出: always, auto, never")
	fmt.Println("  --no-color      禁用颜色输出")
	fmt.Println("  -e FILE -comment \"TEXT\"  设置文件注释")
	fmt.Println("  --help          显示帮助信息")
	fmt.Println("  --version       显示版本信息")
	fmt.Println("")
	fmt.Printf("配置文件 (~/%s):\n", configFileName)
	fmt.Println("  color = always|auto|never")
	fmt.Println("  sort = name|time|size")
	fmt.Println("  group-directories-first = true|false")
	fmt.Println("  human-readable = true|false")
	fmt.Println("  show-hidden = true|false")
	fmt.Println("  time-style = default|iso|long-iso|full-iso")
	fmt.Println("  ignore = PATTERN")
	fmt.Println("")
	fmt.Println("颜色说明:")
	fmt.Println("  蓝色粗体 = 目录")
	fmt.Println("  绿色     = 可执行文件")
	fmt.Println("  青色     = 符号链接")
	fmt.Println("  灰色     = 注释")
	fmt.Println("")
	fmt.Println("平台说明:")
	fmt.Println("  Linux: 注释存储在 xattr 扩展属性中 (user.llc.comment)")
	fmt.Println("  macOS: 注释存储在 Spotlight 元数据中 (kMDItemFinderComment)")
}

func matchesPattern(name, pattern string) bool {
	re := patternCache.get(pattern)
	if re == nil {
		return false
	}
	return re.MatchString(name)
}

func shouldIgnore(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchesPattern(name, pattern) {
			return true
		}
	}
	return false
}

func getComment(path string) string {
	// Try xattr first (works on Linux, also on macOS)
	size, err := unix.Getxattr(path, xattrCommentName, nil)
	if err != nil || size <= 0 {
		return ""
	}

	buf := make([]byte, size)
	size, err = unix.Getxattr(path, xattrCommentName, buf)
	if err != nil || size <= 0 {
		return ""
	}

	return string(buf[:size])
}

func setComment(path, comment string) {
	path = expandPath(path)

	// Check if path contains wildcards
	if strings.Contains(path, "*") || strings.Contains(path, "?") {
		// Batch processing with glob
		matches, err := filepath.Glob(path)
		if err != nil || len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "llc: 没有匹配的文件: %s\n", path)
			os.Exit(1)
		}

		for _, match := range matches {
			setCommentSingle(match, comment)
		}
		fmt.Printf("已设置 %d 个文件的注释\n", len(matches))
		return
	}

	setCommentSingle(path, comment)
}

func setCommentSingle(path, comment string) {
	var err error
	if comment == "" {
		err = unix.Removexattr(path, xattrCommentName)
	} else {
		err = unix.Setxattr(path, xattrCommentName, []byte(comment), 0)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "llc: 设置注释失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已设置注释: [%s] -> %s\n", comment, path)
}

func listDirectory(dirPath string, showAll, showAlmostAll, humanReadable bool, sortBy string, reverseSort bool, showInode, classify, useColor bool, timeStyle string, ignorePatterns []string, groupDirsFirst, followSymlinks bool) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "llc: cannot open directory '%s': %v\n", dirPath, err)
		os.Exit(1)
	}

	var files []FileInfo

	for _, entry := range entries {
		name := entry.Name()

		if showAlmostAll {
			if name == "." || name == ".." {
				continue
			}
		} else if !showAll && strings.HasPrefix(name, ".") {
			continue
		}

		if shouldIgnore(name, ignorePatterns) {
			continue
		}

		fullPath := filepath.Join(dirPath, name)
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:  name,
			Path:  fullPath,
			Info:  info,
			IsDir: info.IsDir(),
		})
	}

	// Add . and .. for -a
	if showAll {
		for _, name := range []string{".", ".."} {
			fullPath := filepath.Join(dirPath, name)
			info, err := os.Lstat(fullPath)
			if err != nil {
				continue
			}
			files = append([]FileInfo{{
				Name:  name,
				Path:  fullPath,
				Info:  info,
				IsDir: info.IsDir(),
			}}, files...)
		}
	}

	// Sort
	sortFiles(files, sortBy, reverseSort, groupDirsFirst)

	// Fetch comments in parallel
	comments := fetchCommentsParallel(files)

	// Print
	for i, file := range files {
		listFile(file.Path, file.Info, comments[i], humanReadable, showInode, classify, useColor, timeStyle, followSymlinks)
	}
}

func sortFiles(files []FileInfo, sortBy string, reverse, groupDirsFirst bool) {
	less := func(i, j int) bool {
		if groupDirsFirst && files[i].IsDir != files[j].IsDir {
			return files[i].IsDir && !files[j].IsDir
		}

		var result bool
		switch sortBy {
		case "time":
			result = files[i].Info.ModTime().After(files[j].Info.ModTime())
		case "size":
			result = files[i].Info.Size() > files[j].Info.Size()
		default: // name
			result = files[i].Name < files[j].Name
		}

		if reverse {
			return !result
		}
		return result
	}

	sort.Slice(files, less)
}

func fetchCommentsParallel(files []FileInfo) []string {
	comments := make([]string, len(files))
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)

	for i, file := range files {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			comment := getComment(path)
			mu.Lock()
			comments[idx] = comment
			mu.Unlock()
		}(i, file.Path)
	}

	wg.Wait()
	return comments
}

func listFile(path string, info os.FileInfo, comment string, humanReadable, showInode, classify, useColor bool, timeStyle string, followSymlinks bool) {
	if info == nil {
		var err error
		info, err = os.Lstat(path)
		if err != nil {
			return
		}
	}

	// Get file info
	mode := info.Mode()
	sys := info.Sys().(*syscall.Stat_t)
	uid := sys.Uid
	gid := sys.Gid

	// Get owner/group names
	owner := strconv.FormatUint(uint64(uid), 10)
	group := strconv.FormatUint(uint64(gid), 10)
	if u, err := user.LookupId(owner); err == nil {
		owner = u.Username
	}
	if g, err := user.LookupGroupId(group); err == nil {
		group = g.Name
	}

	// Format mode
	modeStr := formatMode(mode, path)

	// Format size
	sizeStr := formatSize(info.Size(), humanReadable)

	// Format time
	timeStr := formatTime(info.ModTime(), timeStyle)

	// Get name with color
	name := filepath.Base(path)
	nameColored := name
	if useColor {
		nameColored = colorizeName(path, mode, useColor, followSymlinks)
	}

	// Type indicator
	indicator := ""
	if classify {
		indicator = getTypeIndicator(mode, path)
	}

	// Build output
	var output string
	if showInode {
		output = fmt.Sprintf("%10d ", sys.Ino)
	}

	output += fmt.Sprintf("%s %2d %-8s %-8s %s %s %s%s",
		modeStr,
		sys.Nlink,
		owner,
		group,
		sizeStr,
		timeStr,
		nameColored,
		indicator,
	)

	// Add comment
	if comment != "" {
		if useColor {
			output += fmt.Sprintf("  %s[%s]%s", gray, comment, reset)
		} else {
			output += fmt.Sprintf("  [%s]", comment)
		}
	}

	fmt.Println(output)
}

func formatMode(mode os.FileMode, path string) string {
	var result strings.Builder

	switch {
	case mode&os.ModeDir != 0:
		result.WriteByte('d')
	case mode&os.ModeSymlink != 0:
		result.WriteByte('l')
	default:
		result.WriteByte('-')
	}

	// Owner
	result.WriteByte(ifThen(mode&0o400 != 0, 'r', '-'))
	result.WriteByte(ifThen(mode&0o200 != 0, 'w', '-'))
	result.WriteByte(ifThen(mode&0o100 != 0, 'x', '-'))

	// Group
	result.WriteByte(ifThen(mode&0o040 != 0, 'r', '-'))
	result.WriteByte(ifThen(mode&0o020 != 0, 'w', '-'))
	result.WriteByte(ifThen(mode&0o010 != 0, 'x', '-'))

	// Other
	result.WriteByte(ifThen(mode&0o004 != 0, 'r', '-'))
	result.WriteByte(ifThen(mode&0o002 != 0, 'w', '-'))
	result.WriteByte(ifThen(mode&0o001 != 0, 'x', '-'))

	return result.String()
}

func ifThen(cond bool, a, b byte) byte {
	if cond {
		return a
	}
	return b
}

func formatSize(size int64, human bool) string {
	if !human {
		return fmt.Sprintf("%8d", size)
	}

	units := []string{"B", "K", "M", "G", "T", "P"}
	value := float64(size)
	unitIndex := 0

	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%8dB", size)
	}
	return fmt.Sprintf("%7.1f%s", value, units[unitIndex])
}

func formatTime(t time.Time, style string) string {
	switch style {
	case "iso":
		return t.Format(dateFormatISO)
	case "long-iso":
		return t.Format(dateFormatLongISO)
	case "full-iso":
		return t.Format(dateFormatFullISO)
	default:
		// Default: similar to ls
		if t.Year() == time.Now().Year() {
			return t.Format(dateFormatDefault)
		}
		return t.Format(dateFormatDefaultYear)
	}
}

func colorizeName(path string, mode os.FileMode, useColor bool, followSymlinks bool) string {
	if !useColor {
		return filepath.Base(path)
	}

	name := filepath.Base(path)

	// Handle symlinks
	if followSymlinks && mode&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return red + name + reset
		}

		if !filepath.IsAbs(target) {
			dir := filepath.Dir(path)
			target = filepath.Join(dir, target)
		}

		info, err := os.Stat(target)
		if err != nil {
			return red + name + reset // broken link
		}

		mode = info.Mode()
	}

	switch {
	case mode&os.ModeDir != 0:
		return blue + bold + name + reset
	case mode&os.ModeSymlink != 0:
		return cyan + name + reset
	case mode&0o111 != 0:
		return green + name + reset
	default:
		return name
	}
}

func getTypeIndicator(mode os.FileMode, path string) string {
	switch {
	case mode&os.ModeDir != 0:
		return "/"
	case mode&os.ModeSymlink != 0:
		return "@"
	default:
		// Check for socket and FIFO
		if mode&os.ModeSocket != 0 {
			return "="
		}
		if mode&os.ModeNamedPipe != 0 {
			return "|"
		}
		if mode&0o111 != 0 {
			return "*"
		}
	}
	return ""
}

func getColoredName(path string, useColor, classify, followSymlinks bool) string {
	if !useColor && !classify {
		return filepath.Base(path)
	}

	info, err := os.Lstat(path)
	if err != nil {
		return filepath.Base(path)
	}

	name := filepath.Base(path)
	mode := info.Mode()

	// Resolve symlink if needed
	if followSymlinks && mode&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err == nil {
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(path), target)
			}
			if info2, err := os.Stat(target); err == nil {
				mode = info2.Mode()
			}
		}
	}

	if useColor {
		name = colorizeName(path, info.Mode(), useColor, followSymlinks)
	}

	if classify {
		name += getTypeIndicator(mode, path)
	}

	return name
}

func listSingleColumn(dirPath string, showAll, showAlmostAll bool, sortBy string, reverseSort, classify, useColor bool, ignorePatterns []string, groupDirsFirst, followSymlinks bool) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "llc: cannot open directory '%s': %v\n", dirPath, err)
		os.Exit(1)
	}

	var files []FileInfo

	for _, entry := range entries {
		name := entry.Name()

		if showAlmostAll {
			if name == "." || name == ".." {
				continue
			}
		} else if !showAll && strings.HasPrefix(name, ".") {
			continue
		}

		if shouldIgnore(name, ignorePatterns) {
			continue
		}

		fullPath := filepath.Join(dirPath, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:  name,
			Path:  fullPath,
			Info:  info,
			IsDir: info.IsDir(),
		})
	}

	if showAll {
		for _, name := range []string{".", ".."} {
			fullPath := filepath.Join(dirPath, name)
			info, err := os.Lstat(fullPath)
			if err != nil {
				continue
			}
			files = append([]FileInfo{{
				Name:  name,
				Path:  fullPath,
				Info:  info,
				IsDir: info.IsDir(),
			}}, files...)
		}
	}

	sortFiles(files, sortBy, reverseSort, groupDirsFirst)

	for _, file := range files {
		fmt.Println(getColoredName(file.Path, useColor, classify, followSymlinks))
	}
}

func listRecursive(dirPath string, showAll, showAlmostAll, humanReadable bool, sortBy string, reverseSort bool, showInode, classify, useColor bool, timeStyle string, ignorePatterns []string, groupDirsFirst, singleColumn, followSymlinks bool, depth int, visited map[string]bool) {
	// Prevent cycles
	absPath, _ := filepath.Abs(dirPath)
	if visited[absPath] || depth > maxRecursionDepth {
		return
	}
	visited[absPath] = true

	if depth > 0 {
		fmt.Println()
	}
	fmt.Printf("%s:\n", dirPath)

	if singleColumn {
		listSingleColumn(dirPath, showAll, showAlmostAll, sortBy, reverseSort, classify, useColor, ignorePatterns, groupDirsFirst, followSymlinks)
	} else {
		listDirectory(dirPath, showAll, showAlmostAll, humanReadable, sortBy, reverseSort, showInode, classify, useColor, timeStyle, ignorePatterns, groupDirsFirst, followSymlinks)
	}

	// Get subdirectories
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	var subdirs []string
	for _, entry := range entries {
		name := entry.Name()

		if showAlmostAll {
			if name == "." || name == ".." {
				continue
			}
		} else if !showAll && strings.HasPrefix(name, ".") {
			continue
		}

		if shouldIgnore(name, ignorePatterns) {
			continue
		}

		if entry.IsDir() {
			subdirs = append(subdirs, filepath.Join(dirPath, name))
		}
	}

	sort.Strings(subdirs)
	if reverseSort {
		for i, j := 0, len(subdirs)-1; i < j; i, j = i+1, j-1 {
			subdirs[i], subdirs[j] = subdirs[j], subdirs[i]
		}
	}

	for _, subdir := range subdirs {
		listRecursive(subdir, showAll, showAlmostAll, humanReadable, sortBy, reverseSort, showInode, classify, useColor, timeStyle, ignorePatterns, groupDirsFirst, singleColumn, followSymlinks, depth+1, visited)
	}
}
