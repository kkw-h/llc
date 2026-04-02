package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// getComment retrieves the comment from xattr
func getComment(path string) string {
	size, err := unix.Getxattr(path, xattrCommentName, nil)
	if err != nil || size <= 0 {
		return ""
	}

	buf := make([]byte, size)
	size, err = unix.Getxattr(path, xattrCommentName, buf)
	if err != nil || size <= 0 {
		return ""
	}

	return string(buf[:size])
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

	fmt.Printf("已设置注释: [%s] -> %s\n", comment, path)
	return nil
}
