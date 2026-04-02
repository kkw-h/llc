package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// TestFormatMode tests the formatMode function
func TestFormatMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected string
	}{
		{
			name:     "regular file",
			mode:     0o644,
			expected: "-rw-r--r--",
		},
		{
			name:     "directory",
			mode:     os.ModeDir | 0o755,
			expected: "drwxr-xr-x",
		},
		{
			name:     "symlink",
			mode:     os.ModeSymlink | 0o777,
			expected: "lrwxrwxrwx",
		},
		{
			name:     "executable",
			mode:     0o755,
			expected: "-rwxr-xr-x",
		},
		{
			name:     "readonly",
			mode:     0o444,
			expected: "-r--r--r--",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMode(tt.mode, "")
			if result != tt.expected {
				t.Errorf("formatMode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestFormatSize tests the formatSize function (non-human readable)
func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "       0"},
		{1024, "    1024"},
		{1048576, " 1048576"},
	}

	for _, tt := range tests {
		result := formatSize(tt.size, false)
		if result != tt.expected {
			t.Errorf("formatSize(%d, false) = %v, want %v", tt.size, result, tt.expected)
		}
	}
}

// TestFormatSizeHumanReadable tests the formatSize function (human readable)
func TestFormatSizeHumanReadable(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "       0B"},           // %8dB = 9 chars
		{512, "     512B"},         // %8dB = 9 chars
		{1024, "    1.0K"},         // %7.1f%s = 8 chars
		{1536, "    1.5K"},         // %7.1f%s = 8 chars
		{1024 * 1024, "    1.0M"},  // %7.1f%s = 8 chars
		{1024 * 1024 * 1024, "    1.0G"}, // %7.1f%s = 8 chars
	}

	for _, tt := range tests {
		result := formatSize(tt.size, true)
		if result != tt.expected {
			t.Errorf("formatSize(%d, true) = %q (len=%d), want %q (len=%d)",
				tt.size, result, len(result), tt.expected, len(tt.expected))
		}
	}
}

// TestFormatTime tests the formatTime function
func TestFormatTime(t *testing.T) {
	now := time.Now()
	lastYear := now.AddDate(-1, 0, 0)

	tests := []struct {
		name     string
		t        time.Time
		style    string
		checkFmt bool
		contains string
	}{
		{
			name:     "default current year",
			t:        now,
			style:    "default",
			checkFmt: true,
			contains: ":", // current year shows time
		},
		{
			name:     "default last year",
			t:        lastYear,
			style:    "default",
			checkFmt: false, // Just check it returns non-empty
		},
		{
			name:     "iso format",
			t:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			style:    "iso",
			checkFmt: false,
		},
		{
			name:     "long-iso format",
			t:        time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
			style:    "long-iso",
			checkFmt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(tt.t, tt.style)
			if tt.checkFmt && tt.contains != "" {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("formatTime() = %v, should contain %v", result, tt.contains)
				}
			}
			// Just verify it doesn't panic and returns non-empty
			if result == "" {
				t.Error("formatTime() returned empty string")
			}
		})
	}
}

// TestExpandPath tests the expandPath function
func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~", "~"}, // Just ~ without slash stays as is
	}

	for _, tt := range tests {
		result := expandPath(tt.input)
		if tt.input == "~/test" && result != tt.expected {
			t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
		// For non-tilde paths, result should equal input
		if !strings.HasPrefix(tt.input, "~/") && result != tt.input {
			t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.input)
		}
	}
}

// TestShouldUseColor tests the shouldUseColor function
func TestShouldUseColor(t *testing.T) {
	// Save original NO_COLOR
	origNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", origNoColor)

	tests := []struct {
		name       string
		colorFlag  string
		noColor    bool
		configColor string
		noColorEnv string
		expected   bool
	}{
		{
			name:      "NO_COLOR set",
			noColorEnv: "1",
			expected:  false,
		},
		{
			name:     "no-color flag",
			noColor:  true,
			expected: false,
		},
		{
			name:      "color always",
			colorFlag: "always",
			expected:  true,
		},
		{
			name:       "color never",
			colorFlag:  "never",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColorEnv)
			result := shouldUseColor(tt.colorFlag, tt.noColor, tt.configColor)
			if result != tt.expected {
				t.Errorf("shouldUseColor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMatchesPattern tests the matchesPattern function
func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		match   bool
	}{
		{"exact match", "test.txt", "test.txt", true},
		{"star wildcard", "*.txt", "file.txt", true},
		{"star no match", "*.txt", "file.go", false},
		{"question mark", "file?.txt", "file1.txt", true},
		{"multiple stars", "*.*", "file.txt", true},
		{"no match", "*.go", "test.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.input, tt.pattern)
			if result != tt.match {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.input, tt.pattern, result, tt.match)
			}
		})
	}
}

// TestShouldIgnore tests the shouldIgnore function
func TestShouldIgnore(t *testing.T) {
	patterns := []string{"*.log", "*.tmp", ".git"}

	tests := []struct {
		name     string
		patterns []string
		filename string
		expected bool
	}{
		{"match log", patterns, "debug.log", true},
		{"match tmp", patterns, "temp.tmp", true},
		{"match git", patterns, ".git", true},
		{"no match", patterns, "main.go", false},
		{"empty patterns", []string{}, "file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnore(tt.filename, tt.patterns)
			if result != tt.expected {
				t.Errorf("shouldIgnore(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestLoadConfig tests the loadConfig function
func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".llcrc")

	content := `# Test config
color = always
sort = time
group-directories-first = true
human-readable = true
show-hidden = true
time-style = iso
ignore = *.log
ignore = *.tmp
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Temporarily change home directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	config := loadConfig()

	if config.Color != "always" {
		t.Errorf("Color = %v, want always", config.Color)
	}
	if config.Sort != "time" {
		t.Errorf("Sort = %v, want time", config.Sort)
	}
	if !config.GroupDirsFirst {
		t.Error("GroupDirsFirst should be true")
	}
	if !config.HumanReadable {
		t.Error("HumanReadable should be true")
	}
	if !config.ShowHidden {
		t.Error("ShowHidden should be true")
	}
	if config.TimeStyle != "iso" {
		t.Errorf("TimeStyle = %v, want iso", config.TimeStyle)
	}
	if len(config.IgnorePatterns) != 2 {
		t.Errorf("IgnorePatterns length = %v, want 2", len(config.IgnorePatterns))
	}
}

// TestGetTypeIndicator tests the getTypeIndicator function
func TestGetTypeIndicator(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected string
	}{
		{"directory", os.ModeDir, "/"},
		{"symlink", os.ModeSymlink, "@"},
		{"executable", 0o755, "*"},
		{"regular file", 0o644, ""},
		{"socket", os.ModeSocket, "="},
		{"pipe", os.ModeNamedPipe, "|"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeIndicator(tt.mode, "")
			if result != tt.expected {
				t.Errorf("getTypeIndicator() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestIfThen tests the ifThen helper function
func TestIfThen(t *testing.T) {
	tests := []struct {
		cond     bool
		trueVal  byte
		falseVal byte
		expected byte
	}{
		{true, 'a', 'b', 'a'},
		{false, 'a', 'b', 'b'},
	}

	for _, tt := range tests {
		result := ifThen(tt.cond, tt.trueVal, tt.falseVal)
		if result != tt.expected {
			t.Errorf("ifThen(%v, %c, %c) = %c, want %c", tt.cond, tt.trueVal, tt.falseVal, result, tt.expected)
		}
	}
}

// TestXattrOperations tests setting and getting xattr comments
func TestXattrOperations(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "llc_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test setting and getting comment
	testComment := "test comment 中文"

	// Set comment
	err = unix.Setxattr(tmpFile.Name(), xattrCommentName, []byte(testComment), 0)
	if err != nil {
		t.Fatalf("Failed to set xattr: %v", err)
	}

	// Get comment
	size, err := unix.Getxattr(tmpFile.Name(), xattrCommentName, nil)
	if err != nil {
		t.Fatalf("Failed to get xattr size: %v", err)
	}
	if size <= 0 {
		t.Fatal("xattr size should be > 0")
	}

	buf := make([]byte, size)
	size, err = unix.Getxattr(tmpFile.Name(), xattrCommentName, buf)
	if err != nil {
		t.Fatalf("Failed to get xattr: %v", err)
	}

	result := string(buf[:size])
	if result != testComment {
		t.Errorf("Got comment %q, want %q", result, testComment)
	}

	// Remove comment
	err = unix.Removexattr(tmpFile.Name(), xattrCommentName)
	if err != nil {
		t.Fatalf("Failed to remove xattr: %v", err)
	}

	// Verify removal
	size, _ = unix.Getxattr(tmpFile.Name(), xattrCommentName, nil)
	if size > 0 {
		t.Error("xattr should be removed")
	}
}

// TestSortFiles tests the file sorting functionality
func TestSortFiles(t *testing.T) {
	files := []FileInfo{
		{Name: "b.txt", IsDir: false, Info: &mockFileInfo{name: "b.txt", size: 100}},
		{Name: "a.txt", IsDir: false, Info: &mockFileInfo{name: "a.txt", size: 200}},
		{Name: "dir", IsDir: true, Info: &mockFileInfo{name: "dir", size: 50}},
	}

	// Test sort by name
	t.Run("sort by name", func(t *testing.T) {
		filesCopy := make([]FileInfo, len(files))
		copy(filesCopy, files)
		sortFiles(filesCopy, "name", false, false)
		if filesCopy[0].Name != "a.txt" {
			t.Errorf("First file should be a.txt, got %s", filesCopy[0].Name)
		}
	})

	// Test sort by size
	t.Run("sort by size", func(t *testing.T) {
		filesCopy := make([]FileInfo, len(files))
		copy(filesCopy, files)
		sortFiles(filesCopy, "size", false, false)
		if filesCopy[0].Name != "a.txt" { // largest
			t.Errorf("First file should be a.txt (largest), got %s", filesCopy[0].Name)
		}
	})

	// Test group dirs first
	t.Run("group directories first", func(t *testing.T) {
		filesCopy := make([]FileInfo, len(files))
		copy(filesCopy, files)
		sortFiles(filesCopy, "name", false, true)
		if !filesCopy[0].IsDir {
			t.Errorf("First item should be directory, got %s", filesCopy[0].Name)
		}
	})

	// Test reverse sort
	t.Run("reverse sort", func(t *testing.T) {
		filesCopy := make([]FileInfo, len(files))
		copy(filesCopy, files)
		sortFiles(filesCopy, "name", true, false)
		if filesCopy[0].Name != "dir" {
			t.Errorf("First file should be dir (reverse), got %s", filesCopy[0].Name)
		}
	})
}

// mockFileInfo is a mock implementation of os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// Benchmark tests

func BenchmarkFormatSize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatSize(1024*1024*1024, false)
	}
}

func BenchmarkFormatSizeHumanReadable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatSize(1024*1024*1024, true)
	}
}

func BenchmarkMatchesPattern(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matchesPattern("file.txt", "*.txt")
	}
}

func BenchmarkShouldIgnore(b *testing.B) {
	patterns := []string{"*.log", "*.tmp", "*.swp", "node_modules", ".git"}
	for i := 0; i < b.N; i++ {
		shouldIgnore("test.log", patterns)
	}
}

// TestListDirectory tests the listDirectory function
func TestListDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []string{"a.txt", "b.go", "c.md"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test basic listing
	t.Run("basic listing", func(t *testing.T) {
		// Just verify it doesn't panic
		err := listDirectory(tmpDir, false, false, false, "name", false, false, false, false, "", []string{}, false, false)
		if err != nil {
			t.Errorf("listDirectory failed: %v", err)
		}
	})
}

// TestListRecursive tests the listRecursive function
func TestListRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create files
	os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("test"), 0644)

	// Test recursive listing
	t.Run("recursive listing", func(t *testing.T) {
		visited := make(map[string]bool)
		err := listRecursive(tmpDir, false, false, false, "name", false, false, false, false, "", []string{}, false, false, false, 0, visited)
		if err != nil {
			t.Errorf("listRecursive failed: %v", err)
		}
	})
}

// TestLoadGitignore tests the loadGitignore function
func TestLoadGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .gitignore file
	gitignoreContent := `# Test gitignore
*.log
*.tmp
node_modules/
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	patterns, err := loadGitignore(tmpDir)
	if err != nil {
		t.Fatalf("loadGitignore failed: %v", err)
	}

	if len(patterns) != 3 {
		t.Errorf("Expected 3 patterns, got %d", len(patterns))
	}

	expected := []string{"*.log", "*.tmp", "node_modules/"}
	for i, exp := range expected {
		if patterns[i] != exp {
			t.Errorf("Pattern %d: expected %q, got %q", i, exp, patterns[i])
		}
	}
}

// TestOutputJSON tests the JSON output function
func TestOutputJSON(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644)

	err := outputJSON(tmpDir, false, false, "name", false, []string{}, false)
	if err != nil {
		t.Errorf("outputJSON failed: %v", err)
	}
}

// TestOutputCSV tests the CSV output function
func TestOutputCSV(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644)

	err := outputCSV(tmpDir, false, false, "name", false, []string{}, false)
	if err != nil {
		t.Errorf("outputCSV failed: %v", err)
	}
}

// TestListTree tests the tree output function
func TestListTree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("test"), 0644)

	err := listTree(tmpDir, false, false, "name", false, []string{}, false, false, "")
	if err != nil {
		t.Errorf("listTree failed: %v", err)
	}
}

// TestCommentCache tests the xattr caching functionality
func TestCommentCache(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "llc_cache_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	testComment := "cached comment"

	// Clear cache before test
	globalCommentCache.clear()

	// Set comment
	if err := unix.Setxattr(tmpFile.Name(), xattrCommentName, []byte(testComment), 0); err != nil {
		t.Skipf("xattr not supported: %v", err)
	}

	// First get - should hit filesystem
	comment1 := getComment(tmpFile.Name())
	if comment1 != testComment {
		t.Errorf("First get: expected %q, got %q", testComment, comment1)
	}

	// Second get - should hit cache
	comment2 := getComment(tmpFile.Name())
	if comment2 != testComment {
		t.Errorf("Second get: expected %q, got %q", testComment, comment2)
	}

	// Verify cache has entry
	if _, found := globalCommentCache.get(tmpFile.Name()); !found {
		t.Error("Expected cache to have entry")
	}

	// Clean up
	unix.Removexattr(tmpFile.Name(), xattrCommentName)
}
