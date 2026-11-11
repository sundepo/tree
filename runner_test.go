package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
)

func TestTreeRunner_Integration(t *testing.T) {
	color.NoColor = true

	tmpRoot := t.TempDir()

	// создаем тестовую структуру файлов и папок:
	// tmpRoot/
	// ├── .hidden_dir/
	// │   └── file_in_hidden.txt
	// ├── docs/
	// │   ├── text1.txt (10 bytes)
	// │   └── text2.txt (100 bytes)
	// ├── .hidden_file.txt
	// └── root_file.txt (50 bytes)

	dirs := []string{
		filepath.Join(tmpRoot, ".hidden_dir"),
		filepath.Join(tmpRoot, "docs"),
	}
	for _, dir := range dirs {
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	baseTime := time.Now()

	createFile := func(path string, size int, modTime time.Time) {
		data := make([]byte, size)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", path, err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("failed to set chtimes for %s: %v", path, err)
		}
	}

	createFile(filepath.Join(tmpRoot, ".hidden_dir", "file_in_hidden.txt"), 10, baseTime)
	createFile(filepath.Join(tmpRoot, "docs", "text1.txt"), 10, baseTime.Add(-3*time.Hour))
	createFile(filepath.Join(tmpRoot, "docs", "text2.txt"), 100, baseTime)
	createFile(filepath.Join(tmpRoot, ".hidden_file.txt"), 5, baseTime)
	createFile(filepath.Join(tmpRoot, "root_file.txt"), 50, baseTime.Add(-2*time.Hour))

	docsTime := baseTime.Add(-1 * time.Hour)
	if err := os.Chtimes(filepath.Join(tmpRoot, "docs"), docsTime, docsTime); err != nil {
		t.Fatalf("failed to set chtimes for docs: %v", err)
	}

	// меняем рабочую директорию на временную, чтобы в выводе были относительные или базовые пути
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current wd: %v", err)
	}
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("failed to change wd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }() // возвращаем wd обратно после теста

	var dirSize int64
	if fi, err := os.Stat(filepath.Join(tmpRoot, "docs")); err == nil {
		dirSize = fi.Size()
	}

	var sizeSortedContent []string
	var reverseSizeSortedContent []string

	if dirSize > 50 {
		// Сценарий для Linux/Unix (папка docs (~4096 байт) тяжелее файла root_file.txt (50 байт))
		sizeSortedContent = []string{
			".",
			"├── docs",
			"│   ├── text2.txt", // 100 байт
			"│   └── text1.txt", // 10 байт
			"└── root_file.txt", // 50 байт
			"",
			"1 directories, 3 files",
		}
		reverseSizeSortedContent = []string{
			".",
			"├── root_file.txt",
			"└── docs",
			"    ├── text1.txt",
			"    └── text2.txt",
			"",
			"1 directories, 3 files",
		}
	} else {
		// Сценарий для Windows (папка docs (0 байт) легче файла root_file.txt (50 байт))
		sizeSortedContent = []string{
			".",
			"├── root_file.txt", // 50 байт
			"└── docs",          // 0 байт
			"    ├── text2.txt", // 100 байт
			"    └── text1.txt", // 10 байт
			"",
			"1 directories, 3 files",
		}
		reverseSizeSortedContent = []string{
			".",
			"├── docs",
			"│   ├── text1.txt",
			"│   └── text2.txt",
			"└── root_file.txt",
			"",
			"1 directories, 3 files",
		}
	}

	tests := []struct {
		name     string
		args     Args
		expected []string
	}{
		{
			name: "Default behavior (Alphabetical, no hidden)",
			args: Args{
				Dirs: []string{"."},
			},
			expected: []string{
				".",
				"├── docs",
				"│   ├── text1.txt",
				"│   └── text2.txt",
				"└── root_file.txt",
				"",
				"1 directories, 3 files",
			},
		},
		{
			name: "With hidden files and directories (-a)",
			args: Args{
				Dirs: []string{"."},
				All:  true,
			},
			expected: []string{
				".",
				"├── .hidden_dir",
				"│   └── file_in_hidden.txt",
				"├── .hidden_file.txt",
				"├── docs",
				"│   ├── text1.txt",
				"│   └── text2.txt",
				"└── root_file.txt",
				"",
				"2 directories, 5 files",
			},
		},
		{
			name: "Directories only (-d)",
			args: Args{
				Dirs:    []string{"."},
				DirOnly: true,
			},
			expected: []string{
				".",
				"└── docs",
				"",
				"1 directories",
			},
		},
		{
			name: "Max depth level 1 (-L 1)",
			args: Args{
				Dirs:  []string{"."},
				Level: 1,
			},
			expected: []string{
				".",
				"├── docs",
				"└── root_file.txt",
				"",
				"1 directories, 1 files",
			},
		},
		{
			name:     "Sort by size (-S)",
			args:     Args{Dirs: []string{"."}, SizeSort: true},
			expected: sizeSortedContent,
		},
		{
			name:     "Sort by size and reverse (-S -r)",
			args:     Args{Dirs: []string{"."}, SizeSort: true, Reverse: true},
			expected: reverseSizeSortedContent,
		},
		{
			name: "Sort by modification time (-t)",
			args: Args{
				Dirs:     []string{"."},
				TimeSort: true,
			},
			expected: []string{
				".",
				"├── docs",
				"│   ├── text2.txt",
				"│   └── text1.txt",
				"└── root_file.txt",
				"",
				"1 directories, 3 files",
			},
		},
		{
			name: "Sort by modification time and reverse (-t -r)",
			args: Args{
				Dirs:     []string{"."},
				TimeSort: true,
				Reverse:  true,
			},
			expected: []string{
				".",
				"├── root_file.txt",
				"└── docs",
				"    ├── text1.txt",
				"    └── text2.txt",
				"",
				"1 directories, 3 files",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			runner := NewTreeRunner(tt.args, &buf)
			err := runner.Run()
			if err != nil {
				t.Fatalf("runner.Run() failed: %v", err)
			}

			outputLines := strings.Split(strings.TrimSpace(buf.String()), "\n")

			var expectedCleaned []string
			for _, exp := range tt.expected {
				if exp != "" || len(expectedCleaned) > 0 { // пропускаем начальные пустые, если будут
					expectedCleaned = append(expectedCleaned, strings.TrimSpace(exp))
				}
			}

			if len(outputLines) != len(expectedCleaned) {
				t.Errorf("expected %d lines, got %d. Output:\n%s", len(expectedCleaned), len(outputLines), buf.String())
				return
			}

			for i := range outputLines {
				got := strings.TrimSpace(outputLines[i])
				want := expectedCleaned[i]
				if got != want {
					t.Errorf("line %d mismatch:\ngot : %q\nwant: %q", i, got, want)
				}
			}
		})
	}
}
