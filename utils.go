package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return path
}

// shouldUseColor determines if color output should be used
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
	if term == "dumb" || term == "" {
		return false
	}

	if fi, err := os.Stdout.Stat(); err == nil {
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// matchesPattern checks if name matches a glob pattern
func matchesPattern(name, pattern string) bool {
	re := patternCache.get(pattern)
	if re == nil {
		return false
	}
	return re.MatchString(name)
}

// shouldIgnore checks if name matches any ignore pattern
func shouldIgnore(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchesPattern(name, pattern) {
			return true
		}
	}
	return false
}

// getCommentColorCode returns the ANSI color code for comments.
// Default is plain text to avoid poor contrast against unknown terminal themes.
func getCommentColorCode() string {
	switch strings.ToLower(os.Getenv("LLC_COMMENT_COLOR")) {
	case "", "none", "plain", "default":
		return ""
	case "gray", "grey":
		return gray
	case "red":
		return red
	case "green":
		return green
	case "yellow":
		return yellow
	case "blue":
		return blue
	case "cyan":
		return cyan
	default:
		return ""
	}
}

// getTypeIndicator returns the type indicator character for -F flag
func getTypeIndicator(mode os.FileMode, path string) string {
	switch {
	case mode&os.ModeDir != 0:
		return "/"
	case mode&os.ModeSymlink != 0:
		return "@"
	default:
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
