package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

type TreeRunner struct {
	args   Args        // аргументы командной строки
	out    io.Writer   // поток вывода
	filter EntryFilter // цепочка фильтров
	sorter EntrySorter // сортировщик

	fileCount int // счетчик файлов
	dirCount  int // счетчик директорий

	dirColor *color.Color
	errColor *color.Color
}

func NewTreeRunner(args Args, out io.Writer) *TreeRunner {
	dirColor := color.New(color.FgBlue, color.Bold)
	errColor := color.New(color.FgRed)

	if args.Output != "" {
		dirColor.DisableColor()
		errColor.DisableColor()
	}

	return &TreeRunner{
		args:     args,
		out:      out,
		filter:   NewFilterChain(args),
		sorter:   NewEntrySorter(args),
		dirColor: dirColor,
		errColor: errColor,
	}
}

func (tr *TreeRunner) Run() error {
	for _, root := range tr.args.Dirs {
		tr.dirColor.Fprintln(tr.out, root)

		initialLevels := tr.args.Level
		if initialLevels == 0 {
			initialLevels = -1
		}

		tr.walk(root, "", initialLevels)
	}
	tr.printStatistics()
	return nil
}

func (tr *TreeRunner) printEntry(prefix string, isLast bool, entry os.DirEntry) {
	connector := tr.getConnector(isLast)

	fmt.Fprintf(tr.out, "%s%s", prefix, connector)

	if entry.IsDir() {
		tr.dirColor.Fprintln(tr.out, entry.Name())
	} else {
		fmt.Fprintln(tr.out, entry.Name())
	}
}

func (tr *TreeRunner) printError(prefix string, err error) {
	fmt.Fprintf(tr.out, "%s   ", prefix)
	tr.errColor.Fprintf(tr.out, "[reading error: %v]\n", err)
}

func (tr *TreeRunner) printStatistics() {
	fmt.Fprintln(tr.out)

	if tr.args.DirOnly {
		fmt.Fprintf(tr.out, "%d directories\n", tr.dirCount)
		return
	}
	fmt.Fprintf(tr.out, "%d directories, %d files\n", tr.dirCount, tr.fileCount)
}

// getConnector возвращает графический символ ветки в зависимости от того, последний ли элемент.
func (tr *TreeRunner) getConnector(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}

// getChildPrefix формирует отступ для вложенных элементов следующего уровня.
func (tr *TreeRunner) getChildPrefix(currentPrefix string, isLast bool) string {
	if isLast {
		return currentPrefix + "    "
	}
	return currentPrefix + "│   "
}

// filterEntries фильтрует список файлов и папок на основе переданных флагов.
func (tr *TreeRunner) filterEntries(entries []os.DirEntry) []os.DirEntry {
	var filtered []os.DirEntry
	for _, entry := range entries {
		if tr.filter.Match(entry) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// sortEntries - функция-заглушка для кастомной сортировки на будущее.
func (tr *TreeRunner) sortEntries(entries []os.DirEntry) []os.DirEntry {
	return tr.sorter.Sort(entries)
}
