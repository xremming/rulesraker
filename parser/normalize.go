package parser

import (
	"io"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/unicode/norm"
)

var newlineReplacer = strings.NewReplacer("\r\n", "\n", "\r", "\n")

func normalize(r io.Reader) (string, error) {
	utf8BOMDecoded := unicode.UTF8BOM.NewDecoder().Reader(r)
	unicodeNormalized := norm.NFKC.Reader(utf8BOMDecoded)

	text, err := io.ReadAll(unicodeNormalized)
	if err != nil {
		return "", err
	}

	return newlineReplacer.Replace(string(text)), nil
}

func splitSections(text string) []string {
	var (
		sections []string
		section  []string
	)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if len(section) > 0 {
				sections = append(sections, strings.Join(section, "\n"))
				section = section[:0]
			}

			continue
		}

		section = append(section, line)
	}
	if len(section) > 0 {
		sections = append(sections, strings.Join(section, "\n"))
	}

	return sections
}
