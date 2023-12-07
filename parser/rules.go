package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type SectionType string

const (
	Part    SectionType = "Part"
	Chapter SectionType = "Chapter"
	Rule    SectionType = "Rule"
	SubRule SectionType = "SubRule"
)

var (
	sectionNumberRegexp = regexp.MustCompile(`^\d\.$`)
	chapterNumberRegexp = regexp.MustCompile(`^\d{3}\.$`)
	ruleNumberRegexp    = regexp.MustCompile(`^\d{3}\.\d+$`)
	subRuleNumberRegexp = regexp.MustCompile(`^\d{3}\.\d+\w+$`)
)

func parsePartType(number string) (SectionType, error) {
	if sectionNumberRegexp.MatchString(number) {
		return Part, nil
	}

	if chapterNumberRegexp.MatchString(number) {
		return Chapter, nil
	}

	if ruleNumberRegexp.MatchString(number) {
		return Rule, nil
	}

	if subRuleNumberRegexp.MatchString(number) {
		return SubRule, nil
	}

	return "", fmt.Errorf("invalid part number %q", number)
}

type Section struct {
	ID       string
	Number   string
	Type     SectionType
	Body     []string
	Examples []string `json:",omitempty"`
}

var numberRegexp = regexp.MustCompile(`^(\d+)(\.((\d+)(\w+)?))?\.?`)

func parseNumber(rule string) (string, string, error) {
	matches := numberRegexp.FindStringSubmatch(rule)
	if matches == nil {
		return "", "", fmt.Errorf("failed to parse rule id from %q", rule)
	}

	ruleWithoutNumber := rule[len(matches[0]):]

	major := matches[1]
	minor := matches[4]
	letter := matches[5]

	if minor == "" && letter == "" {
		return ruleWithoutNumber, fmt.Sprintf("%v.", major), nil
	}

	if letter == "" {
		return ruleWithoutNumber, fmt.Sprintf("%v.%v", major, minor), nil
	}

	return ruleWithoutNumber, fmt.Sprintf("%v.%v%v", major, minor, letter), nil
}

func parseRules(rules []string) ([]Section, error) {
	out := make([]Section, 0, len(rules))

	var err error
	for _, rule := range rules {
		ruleWithoutNumber, number, errParseNumber := parseNumber(rule)
		if errParseNumber != nil {
			err = errors.Join(err, errParseNumber)
			continue
		}

		lines := strings.Split(ruleWithoutNumber, "\n")

		var (
			body     []string
			examples []string
		)

		inBody := true
		for _, line := range lines {
			line = strings.TrimSpace(line)

			isExample := strings.HasPrefix(line, "Example:")
			if isExample {
				inBody = false
			}

			if inBody {
				body = append(body, line)
			} else {
				if !isExample {
					err = errors.Join(err, fmt.Errorf("rule body text after examples %q", rule))
					continue
				}

				example := strings.TrimSpace(strings.TrimPrefix(line, "Example:"))
				examples = append(examples, example)
			}
		}

		partType, errParsePartType := parsePartType(number)
		if err != nil {
			err = errors.Join(err, errParsePartType)
			continue
		}

		if len(body) == 0 {
			err = errors.Join(err, fmt.Errorf("rule with no body text %q", rule))
			continue
		}

		if (partType == Part || partType == Chapter) && len(body) != 1 {
			err = errors.Join(err, fmt.Errorf("rule of type %s must have exactly one body element: %q", partType, rule))
			continue
		}

		out = append(out, Section{
			ID:       number,
			Number:   number,
			Type:     partType,
			Body:     body,
			Examples: examples,
		})
	}

	return out, err
}
