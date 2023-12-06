package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type parsedRules struct {
	effectiveDate time.Time
	rules         []string
	glossary      []string
	credits       []string
}

var effectiveDateRegexp = regexp.MustCompile(`\w+ \d+, \d+`)

type parseState int

const (
	parseStateStart parseState = iota
	parseStateRules
	parseStateGlossary
	parseStateCredits
)

func parseSections(sections []string) (parsedRules, error) {
	var (
		state         parseState
		effectiveDate time.Time
		rules         []string
		glossary      []string
		credits       []string
	)
	for _, section := range sections {
		section = strings.TrimSpace(section)

		if effectiveDate.IsZero() {
			match := effectiveDateRegexp.FindString(section)
			effectiveDate, _ = time.Parse("January 2, 2006", match)
		}

		// The last item of the table of contents is "Credits", after which the rules start.
		if state == parseStateStart && section == "Credits" {
			state = parseStateRules
			continue
		}
		if state == parseStateRules && section == "Glossary" {
			state = parseStateGlossary
			continue
		}
		if state == parseStateGlossary && section == "Credits" {
			state = parseStateCredits
			continue
		}

		if section == "" {
			continue
		}

		switch state {
		case parseStateRules:
			rules = append(rules, section)
		case parseStateGlossary:
			glossary = append(glossary, section)
		case parseStateCredits:
			credits = append(credits, section)
		}
	}

	var err error

	if effectiveDate.IsZero() {
		err = errors.Join(fmt.Errorf("failed to parse effective date from the file"))
	}
	if len(rules) == 0 {
		err = errors.Join(err, fmt.Errorf("failed to parse any rules from the file"))
	}
	if len(glossary) == 0 {
		err = errors.Join(err, fmt.Errorf("failed to parse any glossary items from the file"))
	}
	if len(credits) == 0 {
		err = errors.Join(err, fmt.Errorf("failed to parse any credits from the file"))
	}

	return parsedRules{effectiveDate, rules, glossary, credits}, err
}
