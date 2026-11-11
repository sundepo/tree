package main

import (
	"io/fs"
	"os"
	"testing"
)

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() fs.FileMode          { return 0 }
func (m mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestDirOnlyFilter(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		entry    os.DirEntry
		expected bool
	}{
		{
			name:     "Filter disabled, entry is file",
			enabled:  false,
			entry:    mockDirEntry{name: "file.txt", isDir: false},
			expected: true,
		},
		{
			name:     "Filter disabled, entry is directory",
			enabled:  false,
			entry:    mockDirEntry{name: "src", isDir: true},
			expected: true,
		},
		{
			name:     "Filter enabled, entry is directory",
			enabled:  true,
			entry:    mockDirEntry{name: "src", isDir: true},
			expected: true,
		},
		{
			name:     "Filter enabled, entry is file (should block)",
			enabled:  true,
			entry:    mockDirEntry{name: "file.txt", isDir: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := DirOnlyFilter{enabled: tt.enabled}
			if result := filter.Match(tt.entry); result != tt.expected {
				t.Errorf("DirOnlyFilter(enabled=%v).Match(%s, isDir=%v) = %v; want %v",
					tt.enabled, tt.entry.Name(), tt.entry.IsDir(), result, tt.expected)
			}
		})
	}
}

func TestHiddenFilter(t *testing.T) {
	tests := []struct {
		name       string
		showHidden bool
		entry      os.DirEntry
		expected   bool
	}{
		{
			name:       "Show hidden, regular file",
			showHidden: true,
			entry:      mockDirEntry{name: "file.txt", isDir: false},
			expected:   true,
		},
		{
			name:       "Show hidden, hidden file",
			showHidden: true,
			entry:      mockDirEntry{name: ".gitignore", isDir: false},
			expected:   true,
		},
		{
			name:       "Hide hidden, regular file",
			showHidden: false,
			entry:      mockDirEntry{name: "main.go", isDir: false},
			expected:   true,
		},
		{
			name:       "Hide hidden, hidden file (should block)",
			showHidden: false,
			entry:      mockDirEntry{name: ".git", isDir: true},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := HiddenFilter{showHidden: tt.showHidden}
			if result := filter.Match(tt.entry); result != tt.expected {
				t.Errorf("HiddenFilter(showHidden=%v).Match(%s) = %v; want %v",
					tt.showHidden, tt.entry.Name(), result, tt.expected)
			}
		})
	}
}

func TestChainFilter(t *testing.T) {
	args := Args{DirOnly: true, All: false}
	chain := NewFilterChain(args)

	tests := []struct {
		name     string
		entry    os.DirEntry
		expected bool
	}{
		{
			name:     "Regular directory (passes both)",
			entry:    mockDirEntry{name: "src", isDir: true},
			expected: true,
		},
		{
			name:     "Hidden directory (blocked by HiddenFilter)",
			entry:    mockDirEntry{name: ".git", isDir: true},
			expected: false,
		},
		{
			name:     "Regular file (blocked by DirOnlyFilter)",
			entry:    mockDirEntry{name: "main.go", isDir: false},
			expected: false,
		},
		{
			name:     "Hidden file (blocked by both filters)",
			entry:    mockDirEntry{name: ".env", isDir: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := chain.Match(tt.entry); result != tt.expected {
				t.Errorf("ChainFilter.Match(%s, isDir=%v) = %v; want %v",
					tt.entry.Name(), tt.entry.IsDir(), result, tt.expected)
			}
		})
	}
}

func TestPatternFilter(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		entry    os.DirEntry
		expected bool
	}{
		{
			name:     "Empty pattern allows everything",
			pattern:  "",
			entry:    mockDirEntry{name: "main.go", isDir: false},
			expected: true,
		},
		{
			name:     "Matching file passes",
			pattern:  "*.go",
			entry:    mockDirEntry{name: "main.go", isDir: false},
			expected: true,
		},
		{
			name:     "Non-matching file is blocked",
			pattern:  "*.go",
			entry:    mockDirEntry{name: "README.md", isDir: false},
			expected: false,
		},
		{
			name:     "Directory always passes pattern filter",
			pattern:  "*.go",
			entry:    mockDirEntry{name: "docs", isDir: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := PatternFilter{pattern: tt.pattern}
			if result := filter.Match(tt.entry); result != tt.expected {
				t.Errorf("PatternFilter(pattern=%q).Match(%s, isDir=%v) = %v; want %v",
					tt.pattern, tt.entry.Name(), tt.entry.IsDir(), result, tt.expected)
			}
		})
	}
}

func TestIgnoreFilter(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		entry    os.DirEntry
		expected bool
	}{
		{
			name:     "Empty pattern allows everything",
			pattern:  "",
			entry:    mockDirEntry{name: "main.go", isDir: false},
			expected: true,
		},
		{
			name:     "Matching file is blocked",
			pattern:  "*.tmp",
			entry:    mockDirEntry{name: "cache.tmp", isDir: false},
			expected: false,
		},
		{
			name:     "Matching directory is blocked",
			pattern:  "vendor",
			entry:    mockDirEntry{name: "vendor", isDir: true},
			expected: false,
		},
		{
			name:     "Non-matching file passes",
			pattern:  "*.tmp",
			entry:    mockDirEntry{name: "main.go", isDir: false},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := IgnoreFilter{pattern: tt.pattern}
			if result := filter.Match(tt.entry); result != tt.expected {
				t.Errorf("IgnoreFilter(pattern=%q).Match(%s, isDir=%v) = %v; want %v",
					tt.pattern, tt.entry.Name(), tt.entry.IsDir(), result, tt.expected)
			}
		})
	}
}
