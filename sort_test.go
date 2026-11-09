package main

import (
	"io/fs"
	"os"
	"testing"
	"time"
)

type mockFileInfo struct {
	size    int64
	modTime time.Time
}

func (m mockFileInfo) Name() string       { return "" }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() fs.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() any           { return nil }

type mockSortEntry struct {
	name    string
	size    int64
	modTime time.Time
}

func (m mockSortEntry) Name() string      { return m.name }
func (m mockSortEntry) IsDir() bool       { return false }
func (m mockSortEntry) Type() fs.FileMode { return 0 }
func (m mockSortEntry) Info() (fs.FileInfo, error) {
	return mockFileInfo{size: m.size, modTime: m.modTime}, nil
}

func TestAlphabeticalSorter(t *testing.T) {
	tests := []struct {
		name     string
		input    []os.DirEntry
		expected []string
	}{
		{
			name: "Alphabetical sort",
			input: []os.DirEntry{
				mockSortEntry{name: "c.txt"},
				mockSortEntry{name: "a.txt"},
				mockSortEntry{name: "b.txt"},
			},
			expected: []string{"a.txt", "b.txt", "c.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := AlphabeticalSorter{}
			res := sorter.Sort(tt.input)

			for i, expName := range tt.expected {
				if res[i].Name() != expName {
					t.Errorf("At index %d: got %s, want %s", i, res[i].Name(), expName)
				}
			}
		})
	}
}

func TestSizeSorter(t *testing.T) {
	tests := []struct {
		name     string
		input    []os.DirEntry
		expected []string
	}{
		{
			name: "Different sizes",
			input: []os.DirEntry{
				mockSortEntry{name: "small.txt", size: 100},
				mockSortEntry{name: "big.txt", size: 5000},
				mockSortEntry{name: "medium.txt", size: 500},
			},
			expected: []string{"big.txt", "medium.txt", "small.txt"},
		},
		{
			name: "Same sizes falls back to alphabetical",
			input: []os.DirEntry{
				mockSortEntry{name: "c.txt", size: 200},
				mockSortEntry{name: "a.txt", size: 200},
				mockSortEntry{name: "b.txt", size: 200},
			},
			expected: []string{"a.txt", "b.txt", "c.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := SizeSorter{}
			res := sorter.Sort(tt.input)

			for i, expName := range tt.expected {
				if res[i].Name() != expName {
					t.Errorf("At index %d: got %s, want %s", i, res[i].Name(), expName)
				}
			}
		})
	}
}

func TestTimeSorter(t *testing.T) {
	now := time.Now()
	older := now.Add(-1 * time.Hour)
	oldest := now.Add(-2 * time.Hour)

	tests := []struct {
		name     string
		input    []os.DirEntry
		expected []string
	}{
		{
			name: "Different modification times (newest first)",
			input: []os.DirEntry{
				mockSortEntry{name: "older.txt", modTime: older},
				mockSortEntry{name: "newest.txt", modTime: now},
				mockSortEntry{name: "oldest.txt", modTime: oldest},
			},
			expected: []string{"newest.txt", "older.txt", "oldest.txt"},
		},
		{
			name: "Same modification times falls back to alphabetical",
			input: []os.DirEntry{
				mockSortEntry{name: "c.txt", modTime: now},
				mockSortEntry{name: "a.txt", modTime: now},
				mockSortEntry{name: "b.txt", modTime: now},
			},
			expected: []string{"a.txt", "b.txt", "c.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := TimeSorter{}
			res := sorter.Sort(tt.input)

			for i, expName := range tt.expected {
				if res[i].Name() != expName {
					t.Errorf("At index %d: got %s, want %s", i, res[i].Name(), expName)
				}
			}
		})
	}
}

func TestReverseSorter(t *testing.T) {
	tests := []struct {
		name     string
		input    []os.DirEntry
		expected []string
	}{
		{
			name: "Reverse alphabetical sort",
			input: []os.DirEntry{
				mockSortEntry{name: "a.txt"},
				mockSortEntry{name: "b.txt"},
				mockSortEntry{name: "c.txt"},
			},
			expected: []string{"c.txt", "b.txt", "a.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := ReverseSorter{base: AlphabeticalSorter{}}
			res := sorter.Sort(tt.input)

			for i, expName := range tt.expected {
				if res[i].Name() != expName {
					t.Errorf("At index %d: got %s, want %s", i, res[i].Name(), expName)
				}
			}
		})
	}
}

func TestNewEntrySorter(t *testing.T) {
	input := []os.DirEntry{
		mockSortEntry{name: "small.txt", size: 100},
		mockSortEntry{name: "big.txt", size: 5000},
	}

	tests := []struct {
		name     string
		args     Args
		expected []string
	}{
		{
			name:     "Default (Alphabetical)",
			args:     Args{SizeSort: false, Reverse: false},
			expected: []string{"big.txt", "small.txt"}, // по алфавиту 'b' раньше 's'
		},
		{
			name:     "Size Sort Descending",
			args:     Args{SizeSort: true, Reverse: false},
			expected: []string{"big.txt", "small.txt"}, // от большего к меньшему
		},
		{
			name:     "Reverse Alphabetical",
			args:     Args{SizeSort: false, Reverse: true},
			expected: []string{"small.txt", "big.txt"}, // разворот алфавитного
		},
		{
			name:     "Reverse Size Sort (Smallest to Biggest)",
			args:     Args{SizeSort: true, Reverse: true},
			expected: []string{"small.txt", "big.txt"}, // разворот сортировки по размеру
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputCopy := append([]os.DirEntry(nil), input...)

			sorter := NewEntrySorter(tt.args)
			res := sorter.Sort(inputCopy)

			for i, expName := range tt.expected {
				if res[i].Name() != expName {
					t.Errorf("At index %d: got %s, want %s", i, res[i].Name(), expName)
				}
			}
		})
	}
}
