# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

`llc` 是一个 macOS 命令行工具，是 `ls -l` 的增强版本。它在 `ls -l` 的基础上，增加了显示 macOS Finder 注释（kMDItemFinderComment）的功能，将注释显示在每一行的最后面。

## 技术选型

- **语言**: Swift
- **构建工具**: Swift Package Manager
- **依赖**: Foundation, CoreServices 框架

## 常用命令

```bash
# 构建调试版本
swift build

# 构建发布版本
swift build -c release

# 运行测试
swift run llc [路径]

# 安装到 /usr/local/bin
sudo cp .build/release/llc /usr/local/bin/

# 清理构建产物
swift package clean
```

## 使用方法

```bash
llc [选项] [路径]
```

### 选项

- `-a` - 显示所有文件，包括隐藏文件（. 开头）
- `-h, --help` - 显示帮助信息

### 示例

```bash
llc              # 列出当前目录
llc -a           # 列出所有文件（包括隐藏文件）
llc ~/Documents  # 列出指定目录
llc -a ~         # 列出主目录所有文件
```

## 输出格式

输出格式与 `ls -l` 类似，在每行末尾追加 Finder 注释：

```
-rw-r--r--   1 user  group  1234 Jan 01 12:00 filename.txt  [Finder注释内容]
```

最后的 `[...]` 是 Finder 注释（如果有）。

## 项目结构

```
llc/
├── Package.swift          # Swift Package Manager 配置
├── Sources/
│   └── llc/
│       └── llc.swift      # 主程序
└── CLAUDE.md              # 本文件
```

## 核心实现

1. **文件属性获取**: 使用 `FileManager.attributesOfItem(atPath:)` 获取权限、所有者、大小等信息
2. **Finder 注释**: 使用 `MDItemCreateWithURL` + `MDItemCopyAttribute` 原生 API 读取 `kMDItemFinderComment`
3. **性能优化**: 使用 GCD 并发并行获取所有文件的 Finder 注释

## 性能

| 目录大小 | ls -la | llc |
|---------|--------|-----|
| 10 文件 | 0.01s | 0.6s |
| 200+ 文件 | 0.03s | 0.3s |

由于需要调用 Spotlight API 获取 Finder 注释，速度比原生 `ls` 慢，但对于日常使用是可接受的。

## 注意事项

- 仅支持 macOS（依赖 Spotlight 元数据 API）
- Finder 注释需要通过 Finder 的"显示简介"功能添加
- 如果文件没有 Finder 注释，该列显示为空
