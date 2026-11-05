package main

import (
	"os"
	"slices"
	"strings"
)

// EntrySorter определяет интерфейс для сортировки элементов дерева.
type EntrySorter interface {
	Sort(entry []os.DirEntry) []os.DirEntry
}

// AlphabeticalSorter - базовая стратегия сортировки по алфавиту.
type AlphabeticalSorter struct{}

func (s AlphabeticalSorter) Sort(entries []os.DirEntry) []os.DirEntry {
	slices.SortFunc(entries, func(a, b os.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})
	return entries
}

// SizeSorter - стратегия сортировки по размеру (от большего к меньшему).
type SizeSorter struct{}

func (s SizeSorter) Sort(entries []os.DirEntry) []os.DirEntry {
	slices.SortFunc(entries, func(a, b os.DirEntry) int {
		infoA, errA := a.Info()
		infoB, errB := b.Info()

		var sizeA, sizeB int64
		if errA == nil {
			sizeA = infoA.Size()
		}
		if errB == nil {
			sizeB = infoB.Size()
		}

		if sizeA != sizeB {
			if sizeA > sizeB {
				return -1
			}
			return 1
		}

		return strings.Compare(a.Name(), b.Name())
	})
	return entries
}

// ReverseSorter - декоратор, который разворачивает результат работы другого сортера.
type ReverseSorter struct {
	base EntrySorter
}

func (s ReverseSorter) Sort(entries []os.DirEntry) []os.DirEntry {
	res := s.base.Sort(entries)
	slices.Reverse(res)
	return res
}

// NewEntrySorter создает скомпонованную стратегию сортировки на основе аргументов.
func NewEntrySorter(args Args) EntrySorter {
	var sorter EntrySorter

	// выбор стратегии сортировки
	switch {
	case args.SizeSort:
		sorter = SizeSorter{}
	default:
		sorter = AlphabeticalSorter{}
	}

	// обратный порядок
	if args.Reverse {
		sorter = ReverseSorter{base: sorter}
	}

	return sorter
}
