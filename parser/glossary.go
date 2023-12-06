package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	spaceReplaceRegexp = regexp.MustCompile(` +`)
	allowedRunes       = "0123456789abcdefghijklmnopqrstuvwxyz.-"
)

func glossaryID(s string) string {
	trimmed := strings.ToLower(strings.TrimSpace(s))
	spacesReplaced := spaceReplaceRegexp.ReplaceAllLiteralString(trimmed, "-")

	var out strings.Builder
	for _, c := range spacesReplaced {
		i := strings.IndexRune(allowedRunes, c)
		if i < 0 {
			continue
		}

		out.WriteRune(c)
	}

	return out.String()
}

type GlossaryItem struct {
	ID       string
	KeyText  string
	KeyParts []string
	Body     string
}

func newGlossaryItem(key, body string) GlossaryItem {
	var parts []string
	if key == "Active Player, Nonactive Player Order" {
		parts = append(parts, key)
	} else {
		for _, part := range strings.Split(key, ",") {
			part = strings.TrimSpace(strings.ReplaceAll(part, "(Obsolete)", ""))
			parts = append(parts, part)
		}
	}

	id := glossaryID(parts[0])

	return GlossaryItem{id, key, parts, body}
}

func parseGlossary(items []string) ([]GlossaryItem, error) {
	var (
		err error
		out []GlossaryItem
	)
	for _, item := range items {
		splitted := strings.SplitN(item, "\n", 2)
		if len(splitted) != 2 {
			err = errors.Join(err, fmt.Errorf("glossary item with no body: %q", item))
			continue
		}

		out = append(out, newGlossaryItem(splitted[0], splitted[1]))
	}

	return out, err
}
