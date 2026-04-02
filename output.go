package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// FileInfo holds file information
type FileInfo struct {
	Name    string
	Path    string
	Info    os.FileInfo
	Comment string
	IsDir   bool
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("llc version %s\n", version)
	fmt.Println("Enhanced ls command with file comments support")
	fmt.Println("Platform: Linux/macOS (using xattr/Spotlight)")
}

// printHelp prints help information
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
	fmt.Println("配置文件 (~/.llcrc):")
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

// colorizeName returns the name with color codes
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

// getColoredName returns filename with color and type indicator
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

// listFile prints a single file's details
func listFile(path string, info os.FileInfo, comment string, humanReadable, showInode, classify, useColor bool, timeStyle string, followSymlinks bool) {
	if info == nil {
		var err error
		info, err = os.Lstat(path)
		if err != nil {
			return
		}
	}

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

	modeStr := formatMode(mode, path)
	sizeStr := formatSize(info.Size(), humanReadable)
	timeStr := formatTime(info.ModTime(), timeStyle)

	name := filepath.Base(path)
	nameColored := name
	if useColor {
		nameColored = colorizeName(path, mode, useColor, followSymlinks)
	}

	indicator := ""
	if classify {
		indicator = getTypeIndicator(mode, path)
	}

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

	if comment != "" {
		if useColor {
			output += fmt.Sprintf("  %s[%s]%s", gray, comment, reset)
		} else {
			output += fmt.Sprintf("  [%s]", comment)
		}
	}

	fmt.Println(output)
}

// sortFiles sorts files according to the specified criteria
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

// fetchCommentsParallel fetches comments concurrently
func fetchCommentsParallel(files []FileInfo) []string {
	comments := make([]string, len(files))
	var wg sync.WaitGroup
	var mu sync.Mutex
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

// listDirectory lists files in a directory
func listDirectory(dirPath string, showAll, showAlmostAll, humanReadable bool, sortBy string, reverseSort bool, showInode, classify, useColor bool, timeStyle string, ignorePatterns []string, groupDirsFirst, followSymlinks bool) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "llc: cannot open directory '%s': %v\n", dirPath, err)
		os.Exit(1)
	}

	files := collectFiles(entries, dirPath, showAll, showAlmostAll, ignorePatterns)
	sortFiles(files, sortBy, reverseSort, groupDirsFirst)
	comments := fetchCommentsParallel(files)

	for i, file := range files {
		listFile(file.Path, file.Info, comments[i], humanReadable, showInode, classify, useColor, timeStyle, followSymlinks)
	}
}

// listSingleColumn lists files in a single column
func listSingleColumn(dirPath string, showAll, showAlmostAll bool, sortBy string, reverseSort, classify, useColor bool, ignorePatterns []string, groupDirsFirst, followSymlinks bool) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "llc: cannot open directory '%s': %v\n", dirPath, err)
		os.Exit(1)
	}

	files := collectFiles(entries, dirPath, showAll, showAlmostAll, ignorePatterns)
	sortFiles(files, sortBy, reverseSort, groupDirsFirst)

	for _, file := range files {
		fmt.Println(getColoredName(file.Path, useColor, classify, followSymlinks))
	}
}

// collectFiles collects file information from directory entries
func collectFiles(entries []os.DirEntry, dirPath string, showAll, showAlmostAll bool, ignorePatterns []string) []FileInfo {
	var files []FileInfo

	for _, entry := range entries {
		name := entry.Name()
		isDir := entry.IsDir()

		if !shouldIncludeEntry(name, isDir, showAll, showAlmostAll, ignorePatterns, false) {
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
			IsDir: isDir,
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

	return files
}

// listRecursive lists directory recursively
func listRecursive(dirPath string, showAll, showAlmostAll, humanReadable bool, sortBy string, reverseSort bool, showInode, classify, useColor bool, timeStyle string, ignorePatterns []string, groupDirsFirst, singleColumn, followSymlinks bool, depth int, visited map[string]bool) {
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

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	subdirs := getSubdirs(entries, dirPath, showAll, showAlmostAll, ignorePatterns, sortBy, reverseSort)

	for _, subdir := range subdirs {
		listRecursive(subdir, showAll, showAlmostAll, humanReadable, sortBy, reverseSort, showInode, classify, useColor, timeStyle, ignorePatterns, groupDirsFirst, singleColumn, followSymlinks, depth+1, visited)
	}
}

// shouldIncludeEntry checks if a file entry should be included based on filters
func shouldIncludeEntry(name string, isDir bool, showAll, showAlmostAll bool, ignorePatterns []string, wantDirs bool) bool {
	// Handle special entries
	if name == "." || name == ".." {
		return false
	}

	// Handle hidden files
	if showAlmostAll {
		// Include hidden files except . and ..
	} else if !showAll && strings.HasPrefix(name, ".") {
		return false
	}

	// Check ignore patterns
	if shouldIgnore(name, ignorePatterns) {
		return false
	}

	return true
}

// getSubdirs returns a list of subdirectories (sorted according to sortBy and reverse)
func getSubdirs(entries []os.DirEntry, dirPath string, showAll, showAlmostAll bool, ignorePatterns []string, sortBy string, reverse bool) []string {
	var subdirs []FileInfo

	for _, entry := range entries {
		name := entry.Name()

		if !shouldIncludeEntry(name, entry.IsDir(), showAll, showAlmostAll, ignorePatterns, true) {
			continue
		}

		if entry.IsDir() {
			fullPath := filepath.Join(dirPath, name)
			info, _ := entry.Info()
			subdirs = append(subdirs, FileInfo{
				Name:  name,
				Path:  fullPath,
				Info:  info,
				IsDir: true,
			})
		}
	}

	// Sort subdirs consistently with file listing
	sortFiles(subdirs, sortBy, reverse, false)

	// Extract just the paths
	result := make([]string, len(subdirs))
	for i, d := range subdirs {
		result[i] = d.Path
	}

	return result
}
