package main

import (
	"flag"
	"fmt"
	"os"
)

// Application constants
const (
	// Version
	version = "1.6.0"

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
		if err := setComment(*editComment, commentText); err != nil {
			fmt.Fprintf(os.Stderr, "llc: %v\n", err)
			os.Exit(1)
		}
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
			if err := listRecursive(targetPath, *showAll, *showAlmostAll, *humanReadable, sortBy, *reverseSort, *showInode, *classify, useColor, *timeStyle, ignorePatterns, *groupDirsFirst, *singleColumn, *followSymlinks, 0, make(map[string]bool)); err != nil {
				fmt.Fprintf(os.Stderr, "llc: %v\n", err)
				os.Exit(1)
			}
		} else if *listDirSelf {
			if *singleColumn {
				fmt.Println(getColoredName(targetPath, useColor, *classify, *followSymlinks))
			} else {
				listFile(targetPath, nil, "", *humanReadable, *showInode, *classify, useColor, *timeStyle, *followSymlinks)
			}
		} else if *singleColumn {
			if err := listSingleColumn(targetPath, *showAll, *showAlmostAll, sortBy, *reverseSort, *classify, useColor, ignorePatterns, *groupDirsFirst, *followSymlinks); err != nil {
				fmt.Fprintf(os.Stderr, "llc: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := listDirectory(targetPath, *showAll, *showAlmostAll, *humanReadable, sortBy, *reverseSort, *showInode, *classify, useColor, *timeStyle, ignorePatterns, *groupDirsFirst, *followSymlinks); err != nil {
				fmt.Fprintf(os.Stderr, "llc: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		if *singleColumn {
			fmt.Println(getColoredName(targetPath, useColor, *classify, *followSymlinks))
		} else {
			comment := getComment(targetPath)
			listFile(targetPath, info, comment, *humanReadable, *showInode, *classify, useColor, *timeStyle, *followSymlinks)
		}
	}
}
