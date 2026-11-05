package main

import (
	"os"

	"github.com/alecthomas/kong"
)

func main() {
	var args Args
	kong.Parse(&args, kong.Name("tree"))

	runner := NewTreeRunner(args, os.Stdout)
	_ = runner.Run()
}
