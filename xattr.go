package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// commentCacheEntry represents a cached comment with expiration
type commentCacheEntry struct {
	comment   string
	timestamp time.Time
}

// commentCache provides caching for xattr comments
type commentCache struct {
	mu       sync.RWMutex
	entries  map[string]commentCacheEntry
	ttl      time.Duration
}

var globalCommentCache = &commentCache{
	entries: make(map[string]commentCacheEntry),
	ttl:     5 * time.Second, // Cache TTL for 5 seconds
}

// get retrieves a cached comment if available and not expired
func (c *commentCache) get(path string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[path]
	if !exists {
		return "", false
	}

	if time.Since(entry.timestamp) > c.ttl {
		return "", false
	}

	return entry.comment, true
}

// set stores a comment in the cache
func (c *commentCache) set(path, comment string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[path] = commentCacheEntry{
		comment:   comment,
		timestamp: time.Now(),
	}
}

// clear removes all cached entries
func (c *commentCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]commentCacheEntry)
}

// getComment retrieves the comment from xattr (with caching)
func getComment(path string) string {
	// Try cache first
	if comment, found := globalCommentCache.get(path); found {
		return comment
	}

	size, err := unix.Getxattr(path, xattrCommentName, nil)
	if err != nil || size <= 0 {
		globalCommentCache.set(path, "")
		return ""
	}

	buf := make([]byte, size)
	size, err = unix.Getxattr(path, xattrCommentName, buf)
	if err != nil || size <= 0 {
		globalCommentCache.set(path, "")
		return ""
	}

	comment := string(buf[:size])
	globalCommentCache.set(path, comment)
	return comment
}

// setComment sets a comment on a file or files (supports glob patterns)
func setComment(path, comment string) error {
	path = expandPath(path)

	// Check if path contains wildcards
	if strings.Contains(path, "*") || strings.Contains(path, "?") {
		matches, err := filepath.Glob(path)
		if err != nil || len(matches) == 0 {
			return fmt.Errorf("没有匹配的文件: %s", path)
		}

		for _, match := range matches {
			if err := setCommentSingle(match, comment); err != nil {
				return err
			}
		}
		fmt.Printf("已设置 %d 个文件的注释\n", len(matches))
		return nil
	}

	return setCommentSingle(path, comment)
}

// setCommentSingle sets a comment on a single file
func setCommentSingle(path, comment string) error {
	var err error
	if comment == "" {
		err = unix.Removexattr(path, xattrCommentName)
	} else {
		err = unix.Setxattr(path, xattrCommentName, []byte(comment), 0)
	}

	if err != nil {
		return fmt.Errorf("设置注释失败: %v", err)
	}

	// Update cache after setting
	globalCommentCache.set(path, comment)

	fmt.Printf("已设置注释: [%s] -> %s\n", comment, path)
	return nil
}
