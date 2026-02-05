package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseArchive runs the parser against all .txt files in the archive.
// Only files from 2013-07-11 onwards are tested because older files use a
// different format that the parser doesn't support.
func TestParseArchive(t *testing.T) {
	archiveDir := filepath.Join("..", "archive", "txt")

	// These files are UTF-16 encoded which the parser doesn't support.
	skipFiles := map[string]bool{
		"2016-08-26": true,
		"2016-09-30": true,
		"2022-11-18": true,
	}

	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("Failed to read archive directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No files found in archive directory.")
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".txt" {
			continue
		}

		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			date := strings.TrimSuffix(name, ".txt")
			if date < "2013-07-11" {
				t.Skip("File predates supported format (2013-07-11).")
			}
			if skipFiles[date] {
				t.Skip("File is UTF-16 encoded which the parser doesn't support.")
			}
			path := filepath.Join(archiveDir, name)

			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer f.Close()

			rules, err := Parse(f)
			if err != nil {
				t.Errorf("Failed to parse: %v", err)
			}

			if rules.EffectiveDate.IsZero() {
				t.Error("EffectiveDate is zero.")
			}

			if len(rules.Rules) == 0 {
				t.Error("No rules parsed.")
			}

			if len(rules.Glossary) == 0 {
				t.Error("No glossary items parsed.")
			}
		})
	}
}
