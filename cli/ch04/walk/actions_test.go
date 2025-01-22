package main

import (
	"os"
	"testing"
)

func TestFilterOut(t *testing.T) {
	testsCases := []struct {
		name     string
		file     string
		filename string
		ext      string
		minSize  int64
		expected bool
	}{
		{"FilterNoExtension", "testdata/dir.log", "", "", 0, false},

		{"FilterExtensionMatch", "testdata/dir.log", "", ".log", 0, false},
		{"FilterExtensionNoMatch", "testdata/dir.log", "", ".sh", 0, true},

		{"FilterExtensionSizeMatch", "testdata/dir.log", "", ".log", 10, false},
		{"FilterExtensionSizeNoMatch", "testdata/dir.log", "", ".log", 20, true},

		{"FilterFilenameMatch", "testdata/dir.log", "dir", "", 0, false},
		{"FilterFilenameNoMatch", "testdata/dir.log", "noDir", "", 0, true},
	}

	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			info, err := os.Stat(tc.file)
			if err != nil {
				t.Fatal(err)
			}

			t.Log(tc.file, tc.ext, tc.filename, tc.minSize, info.Name())
			f := filterOut(tc.file, tc.ext, tc.filename, tc.minSize, info)

			if f != tc.expected {
				t.Errorf("Expected '%t', got '%t' instead\n", tc.expected, f)
			}
		})
	}
}
