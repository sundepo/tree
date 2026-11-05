package main

import (
	"os"
	"path/filepath"
	"testing"
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
