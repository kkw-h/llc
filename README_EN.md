# llc

An enhanced `ls -l` command-line tool for macOS and Linux. Displays file **comments** at the end of each line (Spotlight/Finder comments on macOS, xattr extended attributes on Linux).

## Features

- 📋 **Standard ls -l format** - Permissions, owner, size, modification time
- 💬 **File comments** - Display file comments at the end of each line
- 📝 **Set comments** - Set file comments via command line
- 🖥️ **Cross-platform** - Supports macOS and Linux
- 📁 **Hidden files** - `-a`/`-A` options to show all files
- 🎨 **Color output** - Blue for directories, green for executables, cyan for symlinks
- 🔢 **inode display** - `-i` option to show file inode numbers
- 📂 **Directory itself** - `-d` option to list directory itself instead of contents
- 📊 **Smart sorting** - By name, time (`-t`), size (`-S`), with reverse (`-r`)
- 📏 **Human-readable** - `-h` option for KB/MB/GB sizes
- 🔗 **Symlink following** - `-L` option to follow symbolic links
- 🗂️ **Directories first** - `--group-directories-first` to list directories before files
- ⚡ **Concurrent** - Uses goroutines for parallel xattr operations
- ⚙️ **Configuration file** - Supports `~/.llcrc` config file

## Installation

### Quick Install (Recommended)

One-command installation using the install script:

```bash
# Using curl
curl -fsSL https://raw.githubusercontent.com/kkw-h/llc/main/install.sh | bash

# Or using wget
wget -qO- https://raw.githubusercontent.com/kkw-h/llc/main/install.sh | bash
```

The script auto-detects your OS and architecture, downloads the appropriate binary, and installs to `/usr/local/bin/`.

### Download from Release

Download binaries from the [Releases](https://github.com/kkw-h/llc/releases) page:

```bash
# Linux AMD64
curl -L -o llc https://github.com/kkw-h/llc/releases/latest/download/llc-linux-amd64

# Linux ARM64
curl -L -o llc https://github.com/kkw-h/llc/releases/latest/download/llc-linux-arm64

# macOS Intel
curl -L -o llc https://github.com/kkw-h/llc/releases/latest/download/llc-darwin-amd64

# macOS Apple Silicon
curl -L -o llc https://github.com/kkw-h/llc/releases/latest/download/llc-darwin-arm64

# Install
chmod +x llc
sudo mv llc /usr/local/bin/
```

### Verify Installation

```bash
# Check version
llc --version

# Test basic functionality
llc -h
```

### Build from Source

#### Prerequisites

- Go 1.23 or higher

#### Build Steps

```bash
# Clone repository
git clone <repository-url>
cd llc

# Download dependencies
go mod tidy

# Build for current platform
go build -o llc

# Or build for specific platform
GOOS=linux GOARCH=amd64 go build -o llc-linux-amd64

# Install to /usr/local/bin
sudo cp llc /usr/local/bin/

# Verify installation
llc --version
```

## Usage

### Basic Usage

```bash
# List current directory
llc

# List all files (including hidden)
llc -a

# List all files (excluding . and ..)
llc -A

# List specific directory
llc ~/Documents

# List single file
llc file.txt
```

### Set File Comments

```bash
# Set comment for a file
llc -e file.txt "This is an important document"

# Batch set comments (supports wildcards)
llc -e "*.txt" "Text file notes"

# View after setting
llc file.txt
```

### Sorting and Filtering

```bash
# Sort by time (newest first)
llc -t

# Sort by size (largest first)
llc -S

# Reverse sort
llc -r

# Directories before files
llc --group-directories-first

# Ignore specific files
llc --ignore="*.log"

# Recursive listing
llc -R
```

### Output Format

```bash
# Human-readable sizes
llc -h

# Show inode numbers
llc -i

# Add type indicators
llc -F

# Single column output
llc -1

# Time format
llc --time-style=iso        # 2024-01-15
llc --time-style=long-iso   # 2024-01-15 14:30
llc --time-style=full-iso   # 2024-01-15 14:30:00 +0800
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR=1` | Disable color output |

### Command Line Options

| Option | Description |
|--------|-------------|
| `-a` | Show all files including hidden (including `.` and `..`) |
| `-A` | Show all files including hidden (excluding `.` and `..`) |
| `-1` | Single column output (one filename per line) |
| `-i` | Show file inode numbers |
| `-d` | List directory itself instead of contents |
| `-h` | Human-readable file sizes (KB, MB, GB) |
| `-F` | Add type indicators (`/` dir, `*` executable, `@` symlink, `=` socket, `\|` FIFO) |
| `-L` | Follow symbolic links, show target file info |
| `-t` | Sort by modification time (newest first) |
| `-S` | Sort by file size (largest first) |
| `-r` | Reverse sort order |
| `-R` | List subdirectories recursively |
| `--group-directories-first` | List directories before files |
| `--ignore=PATTERN` | Ignore matching files (supports `*` and `?` wildcards) |
| `--time-style=STYLE` | Time format: default, iso, long-iso, full-iso |
| `--color=WHEN` | Color output: always, auto, never |
| `--no-color` | Disable color output |
| `-e FILE "comment"` | Set file comment |
| `--help` | Show help information |
| `--version` | Show version information |

### Note

Go's flag package doesn't support combined options. Use separate flags:

```bash
# ✅ Correct
llc -l -h
llc -l -t -r

# ❌ Not supported
llc -lh
llc -ltr
```

## Configuration File

Supports `~/.llcrc` configuration file:

```ini
# Color setting: always, auto, never
color = auto

# Sort method: name, time, size
sort = name

# Directories before files
group-directories-first = true

# Human-readable sizes
human-readable = true

# Show hidden files
show-hidden = false

# Time format: default, iso, long-iso, full-iso
time-style = default

# Ignore patterns (can be used multiple times)
ignore = *.log
ignore = *.tmp
```

## Output Example

```bash
$ llc -h

-rw-r--r--   1 kkw      kkw          3.8K Apr 02 16:02 README.md  [Project documentation]
-rw-r--r--   1 kkw      kkw          1.5K Apr 02 16:39 config.go
-rw-r--r--   1 kkw      kkw          1.8K Apr 02 16:44 format.go
drwxr-xr-x   3 kkw      kkw          4.0K Apr 02 15:35 Sources
-rwxrwxr-x   1 kkw      kkw          3.0M Apr 02 18:11 llc*
```

Output format matches `ls -l` with comments appended at the end `[...]`.

## How to Add File Comments

### Method 1: Using llc command (Recommended)

```bash
# Set comment for single file
llc -e "filename" "Your comment here"

# Batch set (using wildcards)
llc -e "*.txt" "Text file notes"
```

### Method 2: Using System Tools

**macOS:**
```bash
# Using Finder (GUI)
# 1. Select file → Right click → "Get Info" (Cmd + I)
# 2. Enter content in "Comments" field

# Or use xattr command
xattr -w user.llc.comment "Comment content" filename
```

**Linux:**
```bash
# Using xattr command
setfattr -n user.llc.comment -v "Comment content" filename

# Or using Python xattr library
python3 -c "import xattr; xattr.set('filename', 'user.llc.comment', b'Comment content')"
```

## Technical Implementation

- **Language**: Go 1.23+
- **Build Tool**: Go Modules
- **Core Dependencies**:
  - `golang.org/x/sys/unix` - xattr operations
- **Core Implementation**:
  - `os.ReadDir` / `os.Lstat` - Get file attributes
  - `unix.Getxattr` / `unix.Setxattr` - Read/write comments (Linux and macOS)
  - `goroutine` + `sync.WaitGroup` - Concurrent comment fetching

### Performance

| Directory Size | ls -la | llc v1.x (Swift) | llc v2.x (Go) |
|---------|--------|------------------|---------------|
| 10 files | 0.01s | ~0.6s | ~0.003s |
| 200+ files | 0.03s | ~0.3s | ~0.006s |
| 500 files | 0.04s | - | ~0.01s |

*Go version is 50-200x faster than Swift version, close to native `ls` speed.*

## Project Structure

```
llc/
├── main.go          # Main program entry
├── config.go        # Configuration loading
├── format.go        # Formatting functions
├── output.go        # Output related
├── utils.go         # Utility functions and regex cache
├── xattr.go         # xattr operations
├── main_test.go     # Test files
├── install.sh       # Quick install script
├── go.mod           # Go module configuration
├── CLAUDE.md        # Project development docs
└── README.md        # This file
```

## Development

```bash
# Download dependencies
go mod tidy

# Run tests
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem

# Build
go build -o llc

# Install to system
sudo cp llc /usr/local/bin/
```

## Cross-Platform Builds

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o llc-linux-amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o llc-linux-arm64

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o llc-darwin-amd64

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o llc-darwin-arm64

# Windows (experimental)
GOOS=windows GOARCH=amd64 go build -o llc-windows-amd64.exe
```

## Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/llc

# Remove config file (optional)
rm ~/.llcrc
```

## Notes

- **macOS**: Comments stored in xattr (extended attributes), compatible with Finder comments
- **Linux**: Comments stored in xattr (`user.llc.comment`)
- **Cross-platform**: macOS and Linux comment systems are independent, not automatically synced
- **File system**: xattr requires file system support (ext4, xfs, btrfs, APFS, etc.)
- **Permissions**: Some system files may not have permission to modify xattr

## Changelog

### v2.0.0 (2024-04-02)

- 🎉 **Complete rewrite to Go**
- 🐧 **Added Linux support**
- ⚡ **50-200x performance improvement**
- 📦 **Single static binary**, no dependencies
- 🔧 **Added configuration file** support (`~/.llcrc`)
- 🆕 **New options**: `-A`, `-1`, `-L`, `--ignore`, `--time-style`
- ♻️ **Unified error handling**
- 🧪 **Comprehensive test coverage**

### v1.5.0 and earlier

- Swift version, macOS only

## License

MIT License
