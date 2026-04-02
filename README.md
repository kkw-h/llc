# llc

一个增强版 `ls -l` 命令行工具，支持 macOS 和 Linux。在保留 `ls -l` 标准输出的基础上，额外显示文件的**注释**（macOS 使用 Spotlight/Finder 注释，Linux 使用 xattr 扩展属性）。

## 功能特性

- 📋 **标准 ls -l 格式** - 权限、所有者、大小、修改时间
- 💬 **文件注释显示** - 在每行末尾显示文件注释
- 📝 **设置注释** - 通过命令行直接设置文件注释
- 🖥️ **跨平台支持** - 支持 macOS 和 Linux
- 📁 **隐藏文件支持** - `-a`/`-A` 选项显示所有文件
- 🎨 **颜色输出** - 目录蓝色、可执行文件绿色、符号链接青色
- 🔢 **inode 显示** - `-i` 选项显示文件 inode 号
- 📂 **目录本身** - `-d` 选项列出目录本身而非其内容
- 📊 **智能排序** - 按名称、时间 (`-t`)、大小 (`-S`) 排序，支持反向 (`-r`)
- 📏 **人类可读** - `-h` 选项以 KB/MB/GB 显示文件大小
- 🔗 **符号链接** - `-L` 选项跟随符号链接
- 🗂️ **目录优先** - `--group-directories-first` 目录排在文件前面
- ⚡ **并发优化** - 使用 goroutine 并行获取文件注释，提升大目录性能
- ⚙️ **配置文件** - 支持 `~/.llcrc` 配置文件

## 安装

### 快速安装（推荐）

使用一键安装脚本自动下载并安装最新版本：

```bash
# 使用 curl
curl -fsSL https://raw.githubusercontent.com/kkw-h/llc/main/install.sh | bash

# 或使用 wget
wget -qO- https://raw.githubusercontent.com/kkw-h/llc/main/install.sh | bash
```

脚本会自动检测您的操作系统和架构，下载对应的二进制文件并安装到 `/usr/local/bin/`。

### 从 Release 下载

从 [Releases](https://github.com/kkw-h/llc/releases) 页面下载对应平台的二进制文件：

```bash
# Linux AMD64
curl -L -o llc https://github.com/kkw-h/llc/releases/latest/download/llc-linux-amd64

# Linux ARM64
curl -L -o llc https://github.com/kkw-h/llc/releases/latest/download/llc-linux-arm64

# macOS Intel
curl -L -o llc https://github.com/kkw-h/llc/releases/latest/download/llc-darwin-amd64

# macOS Apple Silicon
curl -L -o llc https://github.com/kkw-h/llc/releases/latest/download/llc-darwin-arm64

# 安装
chmod +x llc
sudo mv llc /usr/local/bin/
```

### 验证安装

```bash
# 检查版本
llc --version

# 测试基本功能
llc -h
```

#### 前提条件

- Go 1.23 或更高版本

#### 构建步骤

```bash
# 克隆仓库
git clone <repository-url>
cd llc

# 下载依赖
go mod tidy

# 构建当前平台
go build -o llc

# 或构建指定平台
GOOS=linux GOARCH=amd64 go build -o llc-linux-amd64
GOOS=darwin GOARCH=arm64 go build -o llc-darwin-arm64

# 安装到 /usr/local/bin
sudo cp llc /usr/local/bin/

# 验证安装
llc --version
```

## 使用方法

### 基本用法

```bash
# 列出当前目录
llc

# 列出所有文件（包括隐藏文件）
llc -a

# 列出所有文件（不包括 . 和 ..）
llc -A

# 列出指定目录
llc ~/Documents

# 列出单个文件
llc file.txt
```

### 设置文件注释

```bash
# 设置文件注释
llc -e file.txt "这是重要文档"

# 批量设置注释（支持通配符）
llc -e "*.txt" "文本文件"

# 设置完成后查看
llc file.txt
```

### 排序和过滤

```bash
# 按时间排序（最新的在前）
llc -t

# 按大小排序（最大的在前）
llc -S

# 反向排序
llc -r

# 目录排在文件前面
llc --group-directories-first

# 忽略特定文件
llc --ignore="*.log"

# 递归列出
llc -R
```

### 输出格式

```bash
# 人类可读大小
llc -h

# 显示 inode 号
llc -i

# 添加类型指示符
llc -F

# 单列输出
llc -1

# 时间格式
llc --time-style=iso        # 2024-01-15
llc --time-style=long-iso   # 2024-01-15 14:30
llc --time-style=full-iso   # 2024-01-15 14:30:00 +0800
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `NO_COLOR=1` | 禁用颜色输出 |

### 命令行选项

| 选项 | 说明 |
|------|------|
| `-a` | 显示所有文件，包括隐藏文件（`.` 开头），包括 `.` 和 `..` |
| `-A` | 显示所有文件，包括隐藏文件，不包括 `.` 和 `..` |
| `-1` | 单列输出（每行一个文件名） |
| `-i` | 显示文件的 inode 号 |
| `-d` | 列出目录本身，而非其内容 |
| `-h` | 以人类可读格式显示文件大小 (KB, MB, GB) |
| `-F` | 在文件名后添加类型指示符 (`/` 目录, `*` 可执行, `@` 符号链接, `=` socket, `\|` FIFO) |
| `-L` | 跟随符号链接，显示目标文件信息 |
| `-t` | 按修改时间排序（最新的在前） |
| `-S` | 按文件大小排序（最大的在前） |
| `-r` | 反向排序 |
| `-R` | 递归列出子目录 |
| `--group-directories-first` | 目录排在文件前面 |
| `--ignore=PATTERN` | 忽略匹配的文件（支持 `*` 和 `?` 通配符） |
| `--time-style=STYLE` | 时间显示格式: default, iso, long-iso, full-iso |
| `--color=WHEN` | 颜色输出: always, auto, never |
| `--no-color` | 禁用颜色输出 |
| `-e FILE "备注"` | 设置文件注释 |
| `--help` | 显示帮助信息 |
| `--version` | 显示版本信息 |

### 注意事项

Go 的 flag 包不支持组合选项，请分开使用：

```bash
# ✅ 正确用法
llc -l -h
llc -l -t -r

# ❌ 不支持
llc -lh
llc -ltr
```

## 配置文件

支持 `~/.llcrc` 配置文件，示例：

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

# 忽略模式（可多次使用）
ignore = *.log
ignore = *.tmp
```

## 输出示例

```bash
$ llc -h

-rw-r--r--   1 kkw      kkw          3.8K Apr 02 16:02 README.md  [项目说明文档]
-rw-r--r--   1 kkw      kkw          1.5K Apr 02 16:39 config.go
-rw-r--r--   1 kkw      kkw          1.8K Apr 02 16:44 format.go
drwxr-xr-x   3 kkw      kkw          4.0K Apr 02 15:35 Sources
-rwxrwxr-x   1 kkw      kkw          3.0M Apr 02 18:11 llc*
```

输出格式与 `ls -l` 保持一致，在末尾追加注释 `[...]`。

## 如何添加文件注释

### 方法 1：使用 llc 命令（推荐）

```bash
# 设置单个文件注释
llc -e "文件名" "你的注释内容"

# 批量设置（使用通配符）
llc -e "*.txt" "文本文件备注"
```

### 方法 2：使用系统工具

**macOS:**
```bash
# 使用 Finder（图形界面）
# 1. 选中文件 → 右键 → "显示简介"（Cmd + I）
# 2. 在"备注"栏输入内容

# 或使用 xattr 命令
xattr -w user.llc.comment "注释内容" 文件名
```

**Linux:**
```bash
# 使用 xattr 命令
setfattr -n user.llc.comment -v "注释内容" 文件名

# 或使用 Python xattr 库
python3 -c "import xattr; xattr.set('文件名', 'user.llc.comment', b'注释内容')"
```

## 技术实现

- **语言**：Go 1.23+
- **构建工具**：Go Modules
- **核心依赖**：
  - `golang.org/x/sys/unix` - xattr 操作
- **核心实现**：
  - `os.ReadDir` / `os.Lstat` - 获取文件属性
  - `unix.Getxattr` / `unix.Setxattr` - 读写注释（Linux 和 macOS）
  - `goroutine` + `sync.WaitGroup` - 并发获取注释

### 性能

| 目录大小 | ls -la | llc v1.x (Swift) | llc v2.x (Go) |
|---------|--------|------------------|---------------|
| 10 个文件 | 0.01s | ~0.6s | ~0.003s |
| 200+ 个文件 | 0.03s | ~0.3s | ~0.006s |
| 500 个文件 | 0.04s | - | ~0.01s |

*Go 版本比 Swift 版本快 50-200 倍，接近原生 `ls` 速度。*

## 项目结构

```
llc/
├── main.go          # 主程序入口
├── config.go        # 配置加载
├── format.go        # 格式化函数
├── output.go        # 输出相关
├── utils.go         # 工具函数和正则缓存
├── xattr.go         # xattr 操作
├── main_test.go     # 测试文件
├── go.mod           # Go 模块配置
├── CLAUDE.md        # 项目开发文档
└── README.md        # 本文件
```

## 开发

```bash
# 下载依赖
go mod tidy

# 运行测试
go test -v ./...

# 运行基准测试
go test -bench=. -benchmem

# 构建
go build -o llc

# 安装到系统
sudo cp llc /usr/local/bin/
```

## 跨平台构建

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o llc-linux-amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o llc-linux-arm64

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o llc-darwin-amd64

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o llc-darwin-arm64

# Windows (实验性)
GOOS=windows GOARCH=amd64 go build -o llc-windows-amd64.exe
```

## 注意事项

- **macOS**：注释存储在 xattr（扩展属性）中，与 Finder 注释互通
- **Linux**：注释存储在 xattr (`user.llc.comment`) 中
- **跨平台**：macOS 和 Linux 的注释系统独立，不会自动同步
- **文件系统**：xattr 需要文件系统支持（ext4、xfs、btrfs、APFS 等）
- **权限**：某些系统文件可能没有修改 xattr 的权限

## 更新日志

### v2.0.0 (2024-04-02)

- 🎉 **完整重写到 Go**
- 🐧 **新增 Linux 支持**
- ⚡ **50-200x 性能提升**
- 📦 **单文件静态二进制**，无依赖
- 🔧 **新增配置文件**支持 (`~/.llcrc`)
- 🆕 **新增选项**：`-A`, `-1`, `-L`, `--ignore`, `--time-style`
- ♻️ **统一错误处理**
- 🧪 **全面测试覆盖**

### v1.5.0 及更早

- Swift 版本，仅支持 macOS

## 卸载

```bash
# 删除二进制文件
sudo rm /usr/local/bin/llc

# 删除配置文件（可选）
rm ~/.llcrc
```

## 许可

MIT License
