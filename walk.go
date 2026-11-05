package main

import (
	"os"
	"path/filepath"
)

func (tr *TreeRunner) walk(currentPath, prefix string, remainingLevels int) {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		tr.printError(prefix, err)
		return
	}

	filtered := tr.filterEntries(entries)
	sorted := tr.sortEntries(filtered)

	for i, entry := range sorted {
		isLast := i == len(sorted)-1

		if entry.IsDir() {
			tr.dirCount++
		} else {
			tr.fileCount++
		}

		tr.printEntry(prefix, isLast, entry)

		if entry.IsDir() && (remainingLevels == -1 || remainingLevels > 1) {
			nextPath := filepath.Join(currentPath, entry.Name())
			nextPrefix := tr.getChildPrefix(prefix, isLast)

			nextLevels := remainingLevels
			if remainingLevels != -1 {
				nextLevels--
			}

			tr.walk(nextPath, nextPrefix, nextLevels)
		}
	}
}
