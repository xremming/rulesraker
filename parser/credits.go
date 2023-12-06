package parser

import "fmt"

func parseCredits(credits []string) ([]string, error) {
	if len(credits) == 0 {
		return nil, fmt.Errorf("zero paragraphs of credits found")
	}

	return credits, nil
}
