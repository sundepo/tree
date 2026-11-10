package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Args struct {
	Dirs []string `arg:"" optional:"" name:"dir" help:"Directory list."`

	// listing options
	All     bool `short:"a" group:"Listing options:" help:"All files are listed."`
	DirOnly bool `short:"d" group:"Listing options:" help:"List directories only."`
	Level   int  `short:"L" default:"0" group:"Listing options:" help:"Descend only level directories deep."`

	// sorting options
	Reverse  bool `short:"r" group:"Sorting options:" help:"Reverse the order of the sort."`
	SizeSort bool `short:"S" xor:"sort" group:"Sorting options:" help:"Sort files by size."`
	TimeSort bool `short:"t" xor:"sort" group:"Sorting options:" help:"Sort files by last modification time."`
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
