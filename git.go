package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Git 状态码常量
const (
	gitUntracked = "??"
	gitModified  = "M"
	gitAdded     = "A"
	gitDeleted   = "D"
	gitRenamed   = "R"
	gitCopied    = "C"
	gitUpdated   = "U"
	gitDirty     = "*" // 目录内有修改的自定义标记
)

var (
	gitRootCache   sync.Map // dir -> git root
	gitStatusCache sync.Map // git root -> status map
)

// getGitStatus returns the 2-character git status for a given file or directory
func getGitStatus(path string, isDir bool) string {
	dir := path
	if !isDir {
		dir = filepath.Dir(path)
	}

	gitRoot, err := getGitRoot(dir)
	if err != nil {
		return ""
	}

	var statusMap map[string]string
	if val, ok := gitStatusCache.Load(gitRoot); ok {
		statusMap = val.(map[string]string)
	} else {
		statusMap = computeGitStatusMap(gitRoot)
		gitStatusCache.Store(gitRoot, statusMap)
	}

	// Try the exact path first
	if status, ok := statusMap[path]; ok {
		return status
	}

	// Also check with trailing slash if it's a directory
	if isDir {
		if status, ok := statusMap[path+"/"]; ok {
			return status
		}
	}

	// For untracked files inside an untracked directory, git status --porcelain
	// only reports the directory name (with a trailing slash sometimes, or just the dir).
	// We need to check if any parent directory is in the statusMap as "??" or similar.
	current := path
	for current != gitRoot && current != "/" && current != "." {
		if status, ok := statusMap[current]; ok {
			if isSignificantStatus(status) {
				return status
			}
		}
		// Also check with trailing slash which git sometimes uses for directories
		if status, ok := statusMap[current+"/"]; ok {
			if isSignificantStatus(status) {
				return status
			}
		}
		current = filepath.Dir(current)
	}

	return ""
}

// getGitRoot finds the root directory of the git repository
func getGitRoot(dir string) (string, error) {
	if val, ok := gitRootCache.Load(dir); ok {
		root := val.(string)
		if root == "" {
			return "", fmt.Errorf("not a git repo")
		}
		return root, nil
	}

	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		gitRootCache.Store(dir, "")
		return "", err
	}

	root := strings.TrimSpace(string(out))
	gitRootCache.Store(dir, root)
	return root, nil
}

// computeGitStatusMap runs git status and parses the porcelain output
func computeGitStatusMap(gitRoot string) map[string]string {
	statusMap := make(map[string]string)

	// Add -u to ensure untracked files are reported
	cmd := exec.Command("git", "-C", gitRoot, "status", "--porcelain", "-z", "-u")
	out, err := cmd.Output()
	if err != nil {
		// Git 命令失败（可能不是 git 仓库），返回空 map
		return statusMap
	}
	if len(out) == 0 {
		// 没有状态变更，返回空 map
		return statusMap
	}

	// Output format: XY PATH\0
	// For renames: XY NEW_PATH\0OLD_PATH\0
	parts := bytes.Split(out, []byte{0})
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if len(part) < 4 {
			continue
		}

		status := string(part[0:2])
		path := string(part[3:])

		// In -z format, directories might not have trailing slashes, but sometimes they do.
		// git root is already an absolute path
		absPath := filepath.Join(gitRoot, path)
		if strings.HasSuffix(path, "/") {
			absPath += "/"
		}

		// If path has a trailing slash in the git output, filepath.Join will remove it.
		// We make sure the exact absolute path gets the status
		statusMap[absPath] = status
		// Also store with trailing slash just in case
		statusMap[absPath+"/"] = status

		// When git reports untracked directories, they often have a trailing slash in path.
		// We make sure the exact absolute path of the directory gets the "??" status.
		cleanPath := strings.TrimSuffix(absPath, "/")
		statusMap[cleanPath] = status

		// Propagate to parent directories
		parent := filepath.Dir(cleanPath)
		for parent != gitRoot && parent != "/" && parent != "." {
			if existing, exists := statusMap[parent]; !exists || strings.HasPrefix(existing, gitUntracked) {
				// We use * to indicate a directory has modified/untracked files inside
				if strings.HasPrefix(status, gitUntracked) {
					statusMap[parent] = gitUntracked + " "
				} else {
					statusMap[parent] = gitDirty + " "
				}
			}
			parent = filepath.Dir(parent)
		}
		// Also mark git root if needed
		if parent == gitRoot {
			if existing, exists := statusMap[parent]; !exists || strings.HasPrefix(existing, gitUntracked) {
				if strings.HasPrefix(status, gitUntracked) {
					statusMap[parent] = gitUntracked + " "
				} else {
					statusMap[parent] = gitDirty + " "
				}
			}
		}

		// Skip the old path part if it's a rename
		if len(status) > 0 && (status[0] == gitRenamed[0] || status[0] == gitCopied[0]) {
			i++
		}
	}

	return statusMap
}

// isSignificantStatus checks if a status code is significant enough to display
func isSignificantStatus(status string) bool {
	if len(status) == 0 {
		return false
	}
	// 检查第一列（暂存区）或第二列（工作区）是否有重要状态
	significant := []string{gitUntracked, gitModified, gitAdded, gitDeleted, gitDirty}
	for _, s := range significant {
		if strings.Contains(status, s) {
			return true
		}
	}
	return false
}
