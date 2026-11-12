package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"
)

func TestArgs_AfterApply(t *testing.T) {
	tmpDir := t.TempDir()

	validSubdir := filepath.Join(tmpDir, "valid_dir")
	if err := os.Mkdir(validSubdir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	nonExistentPath := filepath.Join(tmpDir, "ghost_zone")

	tests := []struct {
		name        string
		inputArgs   Args
		expectError bool
		verify      func(t *testing.T, args Args)
	}{
		{
			name:        "Default directory to dot",
			inputArgs:   Args{Dirs: []string{}, Level: 0},
			expectError: false,
			verify: func(t *testing.T, args Args) {
				if len(args.Dirs) != 1 || args.Dirs[0] != "." {
					t.Errorf("expected default dir to be '.', got %v", args.Dirs)
				}
			},
		},
		{
			name:        "Valid directory path clean and pass",
			inputArgs:   Args{Dirs: []string{validSubdir + "/../valid_dir"}, Level: 0},
			expectError: false,
			verify: func(t *testing.T, args Args) {
				// путь должен очиститься с помощью filepath.Clean
				expectedCleaned := filepath.Clean(validSubdir + "/../valid_dir")
				if args.Dirs[0] != expectedCleaned {
					t.Errorf("expected cleaned path %s, got %s", expectedCleaned, args.Dirs[0])
				}
			},
		},
		{
			name:        "Path does not exist error",
			inputArgs:   Args{Dirs: []string{nonExistentPath}, Level: 0},
			expectError: true,
		},
		{
			name:        "Path is a file error",
			inputArgs:   Args{Dirs: []string{testFile}, Level: 0},
			expectError: true,
		},
		{
			name:        "Valid level zero or positive",
			inputArgs:   Args{Dirs: []string{validSubdir}, Level: 3},
			expectError: false,
		},
		{
			name:        "Invalid level negative error",
			inputArgs:   Args{Dirs: []string{validSubdir}, Level: -1},
			expectError: true,
		},
		{
			name:        "Invalid include pattern (-P) error",
			inputArgs:   Args{Dirs: []string{validSubdir}, Pattern: "[invalid"},
			expectError: true,
		},
		{
			name:        "Invalid ignore pattern (-I) error",
			inputArgs:   Args{Dirs: []string{validSubdir}, IgnorePattern: "[invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.inputArgs

			err := args.AfterApply()

			if (err != nil) != tt.expectError {
				t.Fatalf("AfterApply() error = %v, expectError = %v", err, tt.expectError)
			}

			if !tt.expectError && tt.verify != nil {
				tt.verify(t, args)
			}
		})
	}
}

func TestArgs_MutuallyExclusiveFlags(t *testing.T) {
	tests := []struct {
		name        string
		cliArgs     []string
		expectError bool
	}{
		{
			name:        "Only size sort (-S)",
			cliArgs:     []string{"-S"},
			expectError: false,
		},
		{
			name:        "Only time sort (-t)",
			cliArgs:     []string{"-t"},
			expectError: false,
		},
		{
			name:        "Size sort with reverse (-S -r)",
			cliArgs:     []string{"-S", "-r"},
			expectError: false,
		},
		{
			name:        "Time sort with reverse (-t -r)",
			cliArgs:     []string{"-t", "-r"},
			expectError: false,
		},
		{
			name:        "Conflict: size sort and time sort together (-S -t)",
			cliArgs:     []string{"-S", "-t"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args Args

			parser, err := kong.New(&args,
				kong.Name("tree"),
				kong.Exit(func(int) {}),
			)
			if err != nil {
				t.Fatalf("failed to initialize kong parser: %v", err)
			}

			_, err = parser.Parse(tt.cliArgs)

			if (err != nil) != tt.expectError {
				t.Errorf("Parse(%v) error = %v, expectError = %v", tt.cliArgs, err, tt.expectError)
			}
		})
	}
}

func TestArgs_OutputFlag(t *testing.T) {
	tests := []struct {
		name     string
		cliArgs  []string
		expected string
	}{
		{
			name:     "Default empty output",
			cliArgs:  []string{},
			expected: "",
		},
		{
			name:     "Short flag (-o)",
			cliArgs:  []string{"-o", "tree.txt"},
			expected: "tree.txt",
		},
		{
			name:     "Long flag (--output)",
			cliArgs:  []string{"--output=result.txt"},
			expected: "result.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args Args
			parser, err := kong.New(&args,
				kong.Name("tree"),
				kong.Exit(func(int) {}),
			)
			if err != nil {
				t.Fatalf("failed to initialize kong parser: %v", err)
			}

			if _, err := parser.Parse(tt.cliArgs); err != nil {
				t.Fatalf("Parse(%v) unexpected error: %v", tt.cliArgs, err)
			}

			if args.Output != tt.expected {
				t.Errorf("got Output = %q, want %q", args.Output, tt.expected)
			}
		})
	}
}
