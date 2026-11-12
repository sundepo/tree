package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
)

func main() {
	var args Args
	kong.Parse(&args, kong.Name("tree"))

	var out io.Writer = os.Stdout
	if args.Output != "" {
		file, err := os.Create(args.Output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tree: cannot open output file %q: %v\n", args.Output, err)
			os.Exit(1)
		}
		defer file.Close()
		out = file
	}

	runner := NewTreeRunner(args, out)
	if err := runner.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tree: %v\n", err)
		os.Exit(1)
	}
}
