# llc

一个增强版 `ls -l` 命令行工具，为 macOS 设计。在保留 `ls -l` 标准输出的基础上，额外显示文件的 **Finder 注释**（kMDItemFinderComment）。

## 功能特性

- 📋 **标准 ls -l 格式** - 权限、所有者、大小、修改时间
- 💬 **Finder 注释显示** - 在每行末尾显示文件的 Finder 注释
- 📝 **设置注释** - 通过命令行直接设置 Finder 注释
- 📁 **隐藏文件支持** - `-a` 选项显示所有文件（包括隐藏文件）
- 🎨 **颜色输出** - 目录蓝色、可执行文件绿色、符号链接青色
- 🔢 **inode 显示** - `-i` 选项显示文件 inode 号
- 📂 **目录本身** - `-d` 选项列出目录本身而非其内容
- 📊 **智能排序** - 按名称、时间 (`-t`)、大小 (`-S`) 排序，支持反向 (`-r`)
- 📏 **人类可读** - `-h` 选项以 KB/MB/GB 显示文件大小
- ⚡ **并发优化** - 使用 GCD 并行获取文件注释，提升大目录性能

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone <repository-url>
cd llc

# 构建发布版本
swift build -c release

# 安装到 /usr/local/bin
sudo cp .build/release/llc /usr/local/bin/

# 验证安装
llc --help
```

### 环境要求

- macOS 10.14 或更高版本
- Swift 5.0 或更高版本

## 使用方法

### 基本用法

```bash
# 列出当前目录
llc

# 列出所有文件（包括隐藏文件）
llc -a

# 列出指定目录
llc ~/Documents

# 列出单个文件
llc file.txt
```

### 设置 Finder 注释

```bash
# 设置文件注释
llc -e file.txt "这是重要文档"

# 设置完成后查看
llc file.txt
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `NO_COLOR=1` | 禁用颜色输出 |

### 命令行选项

| 选项 | 说明 |
|------|------|
| `-a` | 显示所有文件，包括隐藏文件（`.` 开头） |
| `-i` | 显示文件的 inode 号 |
| `-d` | 列出目录本身，而非其内容 |
| `-h` | 以人类可读格式显示文件大小 (KB, MB, GB) |
| `-F` | 在文件名后添加类型指示符 (*/=@\|) |
| `-t` | 按修改时间排序（最新的在前） |
| `-S` | 按文件大小排序（最大的在前） |
| `-r` | 反向排序 |
| `-R` | 递归列出子目录 |
| `--color` | 强制启用颜色输出 |
| `-e <文件> "备注"` | 设置文件的 Finder 注释 |
| `--help` | 显示帮助信息 |
| `--version` | 显示版本信息 |

### 组合选项

支持类似 `ls` 的组合选项写法：

```bash
llc -li              # 显示 inode
llc -lh              # 人类可读大小
llc -lt              # 按时间排序
llc -ltr             # 按时间反向排序
llc -lhS             # 人类可读 + 按大小排序
```

## 输出示例

```bash
$ llc -a

drwxr-xr-x   9 kkw      staff         288 Apr 02 10:57 .
drwxr-xr-x  12 kkw      staff         384 Apr 02 10:57 ..
-rw-r--r--   1 kkw      staff        2093 Apr 02 10:53 README.md  [项目说明文档]
-rw-r--r--   1 kkw      staff         471 Apr 02 10:45 Package.swift
drwxr-xr-x   3 kkw      staff          96 Apr 02 10:45 Sources
```

输出格式与 `ls -l` 保持一致，在末尾追加 Finder 注释 `[...]`。

## 如何添加 Finder 注释

### 方法 1：使用 llc 命令

```bash
llc -e "文件名" "你的注释内容"
```

### 方法 2：使用 Finder（图形界面）

1. 在 Finder 中选中文件
2. 右键点击 → "显示简介"（或按 `Cmd + I`）
3. 在"备注"栏中输入注释内容

## 技术实现

- **语言**：Swift
- **构建工具**：Swift Package Manager
- **核心 API**：
  - `FileManager` - 获取文件属性
  - `MDItemCreateWithURL` / `MDItemCopyAttribute` - 读取 Finder 注释
  - `osascript` (AppleScript) - 设置 Finder 注释

### 性能

| 目录大小 | ls -la | llc |
|---------|--------|-----|
| 10 个文件 | 0.01s | ~0.6s |
| 200+ 个文件 | 0.03s | ~0.3s |

*注：由于需要调用 Spotlight API，速度比原生 `ls` 慢，但对于日常使用是可接受的。*

## 项目结构

```
llc/
├── Package.swift          # Swift Package Manager 配置
├── Sources/
│   └── llc/
│       └── llc.swift      # 主程序实现
├── CLAUDE.md              # 项目开发文档
└── README.md              # 本文件
```

## 开发

```bash
# 构建调试版本
swift build

# 运行测试
swift run llc [路径]

# 清理构建产物
swift package clean
```

## 注意事项

- 仅支持 macOS（依赖 Spotlight 元数据 API）
- Finder 注释存储在 Spotlight 索引中，可能需要几秒钟才能更新
- 某些系统文件可能没有 Finder 注释权限

## 许可

MIT License
