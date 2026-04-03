package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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
			if status == "??" || status == "A " || status == "AM" || status == "M " || status == "* " {
				return status
			}
		}
		// Also check with trailing slash which git sometimes uses for directories
		if status, ok := statusMap[current+"/"]; ok {
			if status == "??" || status == "A " || status == "AM" || status == "M " || status == "* " {
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
	if err != nil || len(out) == 0 {
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
			if existing, exists := statusMap[parent]; !exists || existing == "??" {
				// We use * to indicate a directory has modified/untracked files inside
				if status == "??" {
					statusMap[parent] = "??"
				} else {
					statusMap[parent] = "* "
				}
			}
			parent = filepath.Dir(parent)
		}
		// Also mark git root if needed
		if parent == gitRoot {
			if existing, exists := statusMap[parent]; !exists || existing == "??" {
				if status == "??" {
					statusMap[parent] = "??"
				} else {
					statusMap[parent] = "* "
				}
			}
		}

		// Skip the old path part if it's a rename
		if status[0] == 'R' || status[0] == 'C' {
			i++
		}
	}

	return statusMap
}
