# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

`llc` 是一个跨平台命令行工具，是 `ls -l` 的增强版本。它在 `ls -l` 的基础上，增加了显示文件注释的功能，将注释显示在每一行的最后面。

- **macOS**: 使用 Spotlight 元数据 (`kMDItemFinderComment`)
- **Linux**: 使用扩展属性 xattr (`user.llc.comment`)

## 技术选型

- **语言**: Go
- **构建工具**: Go Modules
- **依赖**: `golang.org/x/sys` (用于 xattr 操作)

## 项目结构

```
llc/
├── main.go          # 主程序 (Go)
├── go.mod           # Go 模块配置
├── go.sum           # 依赖校验
├── Package.swift    # [已弃用] Swift 版本配置
├── Sources/         # [已弃用] Swift 源码
└── CLAUDE.md        # 本文件
```

## 常用命令

```bash
# 下载依赖
go mod tidy

# 构建当前平台版本
go build -o llc

# 构建 Linux 版本
GOOS=linux GOARCH=amd64 go build -o llc-linux-amd64

# 构建 macOS Intel 版本
GOOS=darwin GOARCH=amd64 go build -o llc-darwin-amd64

# 构建 macOS Apple Silicon 版本
GOOS=darwin GOARCH=arm64 go build -o llc-darwin-arm64

# 安装到 /usr/local/bin
go build -o llc && sudo cp llc /usr/local/bin/

# 运行测试
./llc [路径]
```

## 使用方法

```bash
llc [选项] [路径]
```

### 常用选项

- `-a` - 显示所有文件，包括隐藏文件
- `-A` - 显示所有文件，不包括 `.` 和 `..`
- `-h` - 人类可读的文件大小 (KB, MB, GB)
- `-t` - 按修改时间排序
- `-S` - 按文件大小排序
- `-r` - 反向排序
- `-R` - 递归列出子目录
- `-F` - 添加类型指示符 (*/=@|)
- `-L` - 跟随符号链接
- `-i` - 显示 inode 号
- `--color=always|auto|never` - 控制颜色输出
- `-e 文件 "注释"` - 设置文件注释
- `--ignore=PATTERN` - 忽略匹配的文件

### 示例

```bash
llc                    # 列出当前目录
llc -a                 # 列出所有文件（包括隐藏文件）
llc -lh                # 人类可读大小
llc -lt                # 按时间排序
llc -R ~               # 递归列出主目录
llc -e file.txt "备注"  # 设置文件注释
llc --ignore="*.log"   # 忽略 log 文件
```

## 输出格式

输出格式与 `ls -l` 类似，在每行末尾追加文件注释：

```
-rw-r--r--   1 user  group  1234 Jan 01 12:00 filename.txt  [文件注释内容]
```

最后的 `[...]` 是文件注释（如果有）。

## 配置文件

支持 `~/.llcrc` 配置文件：

```ini
# 颜色设置: always, auto, never
color = auto

# 排序方式: name, time, size
sort = name

# 目录排在文件前面
group-directories-first = true

# 人类可读大小
human-readable = true

# 显示隐藏文件
show-hidden = false

# 时间格式: default, iso, long-iso, full-iso
time-style = default

# 忽略模式 (可多次使用)
ignore = *.log
ignore = *.tmp
```

## 核心实现

1. **文件属性获取**: 使用 Go 标准库 `os.Lstat()` 获取文件信息
2. **注释存储**:
   - Linux: 使用 `unix.Getxattr()` / `unix.Setxattr()` 操作扩展属性
   - macOS: 使用相同的 xattr API（也可通过 Spotlight，当前使用 xattr）
3. **并发优化**: 使用 goroutine 并行获取所有文件的注释

## 跨平台支持

| 功能 | Linux | macOS |
|------|-------|-------|
| 文件列表 | ✓ | ✓ |
| 彩色输出 | ✓ | ✓ |
| 注释存储 | xattr | xattr |
| 人类可读大小 | ✓ | ✓ |
| 符号链接 | ✓ | ✓ |

**注意**: macOS Finder 注释和 Linux xattr 是两套独立系统，注释不会自动同步。

## 性能

| 目录大小 | ls -la | llc |
|---------|--------|-----|
| 10 文件 | 0.01s | 0.02s |
| 200+ 文件 | 0.03s | 0.05s |

Go 版本的并发性能显著优于原 Swift 版本。

## 历史版本

- **v1.x (Swift)**: 原 Swift 实现，仅支持 macOS
- **v2.x (Go)**: 当前 Go 实现，支持 macOS 和 Linux
