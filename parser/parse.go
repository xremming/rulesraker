package parser

import (
	"io"
	"time"
)

type Rules struct {
	EffectiveDate time.Time
	Rules         []Section
	Glossary      []GlossaryItem
	Credits       []string
}

func Parse(r io.Reader) (Rules, error) {
	normalized, err := normalize(r)
	if err != nil {
		return Rules{}, err
	}

	sections := splitSections(normalized)
	parsed, err := parseSections(sections)
	if err != nil {
		return Rules{}, err
	}

	rules, err := parseRules(parsed.rules)
	if err != nil {
		return Rules{}, err
	}

	glossary, err := parseGlossary(parsed.glossary)
	if err != nil {
		return Rules{}, err
	}

	credits, err := parseCredits(parsed.credits)
	if err != nil {
		return Rules{}, err
	}

	return Rules{parsed.effectiveDate, rules, glossary, credits}, nil
}
