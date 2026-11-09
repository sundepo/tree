package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Args struct {
	Dirs []string `arg:"" optional:"" name:"dir" help:"Directory list."`
	// listing options
	All     bool `short:"a" help:"All files are listed."`
	DirOnly bool `short:"d" help:"List directories only."`
	Level   int  `short:"L" default:"0" help:"Descend only level directories deep."`
	// sorting options
	Reverse  bool `short:"r" help:"Reverse the order of the sort."`
	SizeSort bool `short:"S" help:"Sort files by size."`
	TimeSort bool `short:"t" help:"Sort files by last modification time."`
}

func (a *Args) AfterApply() error {
	if len(a.Dirs) == 0 {
		a.Dirs = []string{"."}
	}

	for i, dir := range a.Dirs {
		cleaned := filepath.Clean(dir)
		a.Dirs[i] = cleaned

		info, err := os.Stat(cleaned)
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", dir)
		}
		if err != nil {
			return fmt.Errorf("path validation error %s: %w", dir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", dir)
		}
	}

	if a.Level < 0 {
		return fmt.Errorf("invalid level, must be greater than 0: %d", a.Level)
	}

	return nil
}
