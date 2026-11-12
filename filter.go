package main

import (
	"os"
	"path/filepath"
	"strings"
)

type EntryFilter interface {
	Match(entry os.DirEntry) bool
}

type lowerCaseDirEntry struct {
	os.DirEntry
}

func (e lowerCaseDirEntry) Name() string {
	return strings.ToLower(e.DirEntry.Name())
}

type IgnoreCaseFilter struct {
	base EntryFilter
}

func (f IgnoreCaseFilter) Match(entry os.DirEntry) bool {
	return f.base.Match(lowerCaseDirEntry{DirEntry: entry})
}

type ChainFilter []EntryFilter

func (cf ChainFilter) Match(entry os.DirEntry) bool {
	for _, filter := range cf {
		if !filter.Match(entry) {
			return false
		}
	}
	return true
}

type DirOnlyFilter struct {
	enabled bool
}

func (f DirOnlyFilter) Match(entry os.DirEntry) bool {
	if f.enabled && !entry.IsDir() {
		return false
	}
	return true
}

type HiddenFilter struct {
	showHidden bool
}

func (f HiddenFilter) Match(entry os.DirEntry) bool {
	if !f.showHidden && entry.Name()[0] == '.' {
		return false
	}
	return true
}

type PatternFilter struct {
	pattern string
}

func (f PatternFilter) Match(entry os.DirEntry) bool {
	if f.pattern == "" || entry.IsDir() {
		return true
	}
	matched, err := filepath.Match(f.pattern, entry.Name())
	if err != nil {
		return false
	}
	return matched
}

type IgnoreFilter struct {
	pattern string
}

func (f IgnoreFilter) Match(entry os.DirEntry) bool {
	if f.pattern == "" {
		return true
	}
	matched, err := filepath.Match(f.pattern, entry.Name())
	if err != nil {
		return true
	}
	return !matched
}

func NewFilterChain(args Args) EntryFilter {
	includePat := args.Pattern
	ignorePat := args.IgnorePattern

	if args.IgnoreCase {
		includePat = strings.ToLower(includePat)
		ignorePat = strings.ToLower(ignorePat)
	}

	var patternFilter EntryFilter = PatternFilter{pattern: includePat}
	var ignoreFilter EntryFilter = IgnoreFilter{pattern: ignorePat}

	if args.IgnoreCase {
		patternFilter = IgnoreCaseFilter{base: patternFilter}
		ignoreFilter = IgnoreCaseFilter{base: ignoreFilter}
	}

	return ChainFilter{
		DirOnlyFilter{enabled: args.DirOnly},
		HiddenFilter{showHidden: args.All},
		patternFilter,
		ignoreFilter,
	}
}
