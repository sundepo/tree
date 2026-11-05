package main

import "os"

type EntryFilter interface {
	Match(entry os.DirEntry) bool
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

func NewFilterChain(args Args) EntryFilter {
	return ChainFilter{
		DirOnlyFilter{enabled: args.DirOnly},
		HiddenFilter{showHidden: args.All},
	}
}
