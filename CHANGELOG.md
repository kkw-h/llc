# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.0] - 2025-04-02

### Added
- **Tree output**: New `--tree` flag for tree-like directory display
- **JSON output**: New `--json` flag for JSON format output
- **CSV output**: New `--csv` flag for CSV format output
- **Gitignore support**: New `--gitignore` flag to apply `.gitignore` rules
- **New sorting options**:
  - `-u`: Sort by access time
  - `-U`: Sort by creation time
  - `--sort-ext`: Sort by file extension
- **Comment caching**: Added 5-second TTL cache for xattr comments
- **Memory pool optimization**: Using `sync.Pool` for FileInfo slices
- **Shell completions**: Added bash, zsh, and fish completion scripts
- **Man page**: Added `llc.1` manual page
- **Homebrew tap**: Available via `brew tap kkw-h/llc`

### Changed
- **Code refactoring**: Restructured main.go with Options and AppConfig structs
- **Improved help text**: Updated with new options

## [2.0.2] - 2025-04-02

### Fixed
- Fixed batch comment setting with shell wildcard expansion (e.g., `llc -e *.txt "comment"` without quotes now works correctly)

## [2.0.1] - 2025-04-02

### Added
- English README (README_EN.md)
- Language switch links between README files

### Changed
- Updated GitHub repository description and topics
- Fixed release notes template to use dynamic version

## [2.0.0] - 2025-04-02

### Added
- **Cross-platform support**: Linux and macOS
- **File comments**: Display and set file comments via xattr/Spotlight
- **Standard ls options**: `-a`, `-A`, `-1`, `-i`, `-d`, `-h`, `-F`, `-L`, `-t`, `-S`, `-r`, `-R`
- **Extended options**:
  - `--group-directories-first`: List directories before files
  - `--ignore=PATTERN`: Ignore matching files
  - `--time-style=STYLE`: Custom time formats (iso, long-iso, full-iso)
  - `--color=WHEN`: Color control (always, auto, never)
- **Configuration file**: Support for `~/.llcrc`
- **Concurrent processing**: Goroutines for parallel xattr operations
- **GitHub Actions**: Automated builds for 4 platforms (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64)
- **Install script**: One-command installation via curl/wget

### Changed
- Complete rewrite from Swift to Go
- 50-200x performance improvement over v1.x

### Removed
- Swift implementation (moved to v1.x branch)

## [1.5.0] and earlier

### Added
- Initial Swift implementation
- macOS support only
- Spotlight/Finder comments integration

[Unreleased]: https://github.com/kkw-h/llc/compare/v2.1.0...HEAD
[2.1.0]: https://github.com/kkw-h/llc/compare/v2.0.2...v2.1.0
[2.0.2]: https://github.com/kkw-h/llc/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/kkw-h/llc/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/kkw-h/llc/releases/tag/v2.0.0
