package main

import (
	"flag"
	"fmt"
	"os"
)

// Application constants
const (
	// Version
	version = "2.1.0"

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

// Options holds all command-line options
type Options struct {
	showAll        bool
	showAlmostAll  bool
	singleColumn   bool
	showInode      bool
	listDirSelf    bool
	humanReadable  bool
	classify       bool
	followSymlinks bool
	sortByTime     bool
	sortBySize     bool
	sortByAccess   bool
	sortByCreate   bool
	sortByExt      bool
	reverseSort    bool
	recursive      bool
	groupDirsFirst bool
	useGitignore   bool
	outputJSON     bool
	outputCSV      bool
	outputTree     bool
	colorFlag      string
	noColor        bool
	timeStyle      string
	ignorePattern  string
	editComment    string
	editCommentText string
	showVersion    bool
	showHelp       bool
}

// AppConfig holds the merged configuration (CLI flags + config file)
type AppConfig struct {
	Options
	sortBy          string
	ignorePatterns  []string
	useColor        bool
	targetPath      string
	configSort      string
}

func main() {
	opts := parseFlags()

	if opts.showVersion {
		printVersion()
		return
	}

	if opts.showHelp {
		printHelp()
		return
	}

	config := buildConfig(opts)

	if config.editComment != "" {
		if err := handleEditComment(config); err != nil {
			fmt.Fprintf(os.Stderr, "llc: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := executeList(config); err != nil {
		fmt.Fprintf(os.Stderr, "llc: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Options {
	var opts Options

	flag.BoolVar(&opts.showAll, "a", false, "显示所有文件，包括隐藏文件")
	flag.BoolVar(&opts.showAlmostAll, "A", false, "显示所有文件，不包括 . 和 ..")
	flag.BoolVar(&opts.singleColumn, "1", false, "单列输出")
	flag.BoolVar(&opts.showInode, "i", false, "显示 inode 号")
	flag.BoolVar(&opts.listDirSelf, "d", false, "列出目录本身")
	flag.BoolVar(&opts.humanReadable, "h", false, "人类可读大小")
	flag.BoolVar(&opts.classify, "F", false, "添加类型指示符")
	flag.BoolVar(&opts.followSymlinks, "L", false, "跟随符号链接")
	flag.BoolVar(&opts.sortByTime, "t", false, "按修改时间排序")
	flag.BoolVar(&opts.sortBySize, "S", false, "按大小排序")
	flag.BoolVar(&opts.sortByAccess, "u", false, "按访问时间排序")
	flag.BoolVar(&opts.sortByCreate, "U", false, "按创建时间排序")
	flag.BoolVar(&opts.sortByExt, "sort-ext", false, "按扩展名排序")
	flag.BoolVar(&opts.reverseSort, "r", false, "反向排序")
	flag.BoolVar(&opts.recursive, "R", false, "递归列出")
	flag.BoolVar(&opts.groupDirsFirst, "group-directories-first", false, "目录排在文件前面")
	flag.BoolVar(&opts.useGitignore, "gitignore", false, "使用 .gitignore 规则")
	flag.BoolVar(&opts.outputJSON, "json", false, "JSON 格式输出")
	flag.BoolVar(&opts.outputCSV, "csv", false, "CSV 格式输出")
	flag.BoolVar(&opts.outputTree, "tree", false, "树形输出")
	flag.StringVar(&opts.colorFlag, "color", "", "颜色输出: always, auto, never")
	flag.BoolVar(&opts.noColor, "no-color", false, "禁用颜色")
	flag.StringVar(&opts.timeStyle, "time-style", "", "时间格式: default, iso, long-iso, full-iso")
	flag.StringVar(&opts.ignorePattern, "ignore", "", "忽略模式")
	flag.StringVar(&opts.editComment, "e", "", "设置注释的文件路径")
	flag.StringVar(&opts.editCommentText, "comment", "", "注释内容（与 -e 配合使用）")
	flag.BoolVar(&opts.showVersion, "version", false, "显示版本")
	flag.BoolVar(&opts.showHelp, "help", false, "显示帮助")

	flag.Parse()
	return opts
}

func buildConfig(opts Options) AppConfig {
	cfg := AppConfig{Options: opts}

	// Load config file
	fileConfig := loadConfig()

	// Apply config file defaults (CLI flags override config)
	if !opts.humanReadable && fileConfig.HumanReadable {
		cfg.humanReadable = true
	}
	if !opts.showAll && !opts.showAlmostAll && fileConfig.ShowHidden {
		cfg.showAll = true
	}
	if !opts.groupDirsFirst && fileConfig.GroupDirsFirst {
		cfg.groupDirsFirst = true
	}
	if opts.timeStyle == "" && fileConfig.TimeStyle != "" {
		cfg.timeStyle = fileConfig.TimeStyle
	}
	cfg.configSort = fileConfig.Sort

	// Determine sort method
	cfg.sortBy = "name"
	if opts.sortByTime {
		cfg.sortBy = "time"
	} else if opts.sortBySize {
		cfg.sortBy = "size"
	} else if opts.sortByAccess {
		cfg.sortBy = "access"
	} else if opts.sortByCreate {
		cfg.sortBy = "create"
	} else if opts.sortByExt {
		cfg.sortBy = "ext"
	} else if fileConfig.Sort != "" {
		cfg.sortBy = fileConfig.Sort
	}

	// Determine color setting
	cfg.useColor = shouldUseColor(opts.colorFlag, opts.noColor, fileConfig.Color)

	// Collect ignore patterns
	cfg.ignorePatterns = fileConfig.IgnorePatterns
	if opts.ignorePattern != "" {
		cfg.ignorePatterns = append(cfg.ignorePatterns, opts.ignorePattern)
	}

	// Get target path
	cfg.targetPath = "."
	args := flag.Args()
	if len(args) > 0 {
		cfg.targetPath = args[0]
	}
	cfg.targetPath = expandPath(cfg.targetPath)

	return cfg
}

func handleEditComment(cfg AppConfig) error {
	args := flag.Args()
	commentText := cfg.editCommentText

	// If --comment is not set, use remaining args
	if commentText == "" {
		if len(args) == 0 {
			return fmt.Errorf("-e 需要指定备注内容\n用法: llc -e <文件> \"备注内容\"")
		}
		// Last arg is the comment, all others are additional files
		commentText = args[len(args)-1]
		// Add any middle args to file list (shell may expand wildcards)
		for i := 0; i < len(args)-1; i++ {
			if err := setCommentSingle(args[i], commentText); err != nil {
				fmt.Fprintf(os.Stderr, "llc: %v\n", err)
			}
		}
	}

	// Set comment for the main file (may contain wildcards)
	return setComment(cfg.editComment, commentText)
}

func executeList(cfg AppConfig) error {
	info, err := os.Stat(cfg.targetPath)
	if err != nil {
		return fmt.Errorf("cannot access '%s': %v", cfg.targetPath, err)
	}

	// Handle gitignore if requested
	if cfg.useGitignore {
		gitignorePatterns, err := loadGitignore(cfg.targetPath)
		if err == nil {
			cfg.ignorePatterns = append(cfg.ignorePatterns, gitignorePatterns...)
		}
	}

	if info.IsDir() {
		return handleDirectory(cfg, info)
	}
	return handleFile(cfg, info)
}

func handleDirectory(cfg AppConfig, info os.FileInfo) error {
	if cfg.recursive {
		return listRecursive(cfg.targetPath, cfg.showAll, cfg.showAlmostAll, cfg.humanReadable,
			cfg.sortBy, cfg.reverseSort, cfg.showInode, cfg.classify, cfg.useColor, cfg.timeStyle,
			cfg.ignorePatterns, cfg.groupDirsFirst, cfg.singleColumn, cfg.followSymlinks,
			0, make(map[string]bool))
	}

	if cfg.listDirSelf {
		if cfg.singleColumn {
			fmt.Println(getColoredName(cfg.targetPath, cfg.useColor, cfg.classify, cfg.followSymlinks))
		} else {
			listFile(cfg.targetPath, nil, "", cfg.humanReadable, cfg.showInode, cfg.classify,
				cfg.useColor, cfg.timeStyle, cfg.followSymlinks)
		}
		return nil
	}

	if cfg.outputJSON {
		return outputJSON(cfg.targetPath, cfg.showAll, cfg.showAlmostAll, cfg.sortBy,
			cfg.reverseSort, cfg.ignorePatterns, cfg.groupDirsFirst)
	}

	if cfg.outputCSV {
		return outputCSV(cfg.targetPath, cfg.showAll, cfg.showAlmostAll, cfg.sortBy,
			cfg.reverseSort, cfg.ignorePatterns, cfg.groupDirsFirst)
	}

	if cfg.outputTree {
		fmt.Println(cfg.targetPath)
		return listTree(cfg.targetPath, cfg.showAll, cfg.showAlmostAll, cfg.sortBy,
			cfg.reverseSort, cfg.ignorePatterns, cfg.groupDirsFirst, cfg.followSymlinks, "")
	}

	if cfg.singleColumn {
		return listSingleColumn(cfg.targetPath, cfg.showAll, cfg.showAlmostAll, cfg.sortBy,
			cfg.reverseSort, cfg.classify, cfg.useColor, cfg.ignorePatterns, cfg.groupDirsFirst,
			cfg.followSymlinks)
	}

	return listDirectory(cfg.targetPath, cfg.showAll, cfg.showAlmostAll, cfg.humanReadable,
		cfg.sortBy, cfg.reverseSort, cfg.showInode, cfg.classify, cfg.useColor, cfg.timeStyle,
		cfg.ignorePatterns, cfg.groupDirsFirst, cfg.followSymlinks)
}

func handleFile(cfg AppConfig, info os.FileInfo) error {
	if cfg.singleColumn {
		fmt.Println(getColoredName(cfg.targetPath, cfg.useColor, cfg.classify, cfg.followSymlinks))
		return nil
	}

	comment := getComment(cfg.targetPath)
	listFile(cfg.targetPath, info, comment, cfg.humanReadable, cfg.showInode, cfg.classify,
		cfg.useColor, cfg.timeStyle, cfg.followSymlinks)
	return nil
}
