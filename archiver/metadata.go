package archiver

import (
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"
)

type MissingDate struct {
	Date    JSONDate
	Source  string
	Comment string `json:",omitempty"`
}

type OneOffURL struct {
	Date      JSONDate
	Available bool
	Files     map[string]string
}

type URLFormats struct {
	Available []string
	OneOff    []OneOffURL
}

type PossibleURL struct {
	Date      time.Time
	Format    string
	URL       url.URL
	Available bool
}

func (u URLFormats) PossibleURLs(date time.Time) []PossibleURL {
	var out []PossibleURL

	for _, ext := range []string{"txt", "pdf", "docx"} {
		varReplacer := strings.NewReplacer(
			"{{year}}", date.Format("2006"),
			"{{yearShort}}", date.Format("06"),
			"{{month}}", date.Format("01"),
			"{{day}}", date.Format("02"),
			"{{ext}}", ext,
		)

		for _, urlTemplate := range u.Available {
			replaced := varReplacer.Replace(urlTemplate)
			url, err := url.Parse(replaced)
			if err != nil {
				panic(err)
			}

			out = append(out, PossibleURL{
				Date:      date,
				Format:    ext,
				URL:       *url,
				Available: true,
			})
		}
	}

	for _, oneOff := range u.OneOff {
		if time.Time(oneOff.Date).Equal(date) {
			for ext, oneOffURL := range oneOff.Files {
				parsedURL, err := url.Parse(oneOffURL)
				if err != nil {
					panic(err)
				}

				out = append(out, PossibleURL{
					Date:      date,
					Format:    ext,
					URL:       *parsedURL,
					Available: oneOff.Available,
				})
			}
		}
	}

	return out
}

type ResponseMetadata struct {
	ContentLength int64
	ContentType   string
	LastModified  string
	ETag          string
}

// FoundFile is a file that was not directly downloaded from the URL but was
// found through other means. Such a file might have come from e.g. the
// https://github.com/pit142857/mtg-cr repository or the Internet Archive.
type FoundFile struct {
	Date        JSONDate
	Format      string
	File        string
	OriginalURL *string
	Source      string
	Comment     string
}

type Rule struct {
	Date   JSONDate
	Format string
	File   string
	URL    *string

	ResponseMetadata *ResponseMetadata `json:",omitempty"`
}

type Metadata struct {
	LatestUpdate       JSONDate
	KnownExistingDates []JSONDate
	KnownMissingDates  []MissingDate
	URLFormats         URLFormats
	FoundFiles         []FoundFile
	Rules              []Rule
}

func (m *Metadata) PrepareForEncoding() {
	// sort found files
	sort.Slice(m.FoundFiles, func(i, j int) bool {
		keyI := fmt.Sprintf("%s_%s", m.FoundFiles[i].Date.String(), m.FoundFiles[i].Format)
		keyJ := fmt.Sprintf("%s_%s", m.FoundFiles[j].Date.String(), m.FoundFiles[j].Format)
		return keyI < keyJ
	})

	// add all found files to rules
	for _, foundFile := range m.FoundFiles {
		m.Rules = append(m.Rules, Rule{
			Date:   foundFile.Date,
			Format: foundFile.Format,
			File:   foundFile.File,
		})
	}

	// add all dates from rules to known existing dates
	for _, rule := range m.Rules {
		m.KnownExistingDates = append(m.KnownExistingDates, rule.Date)
	}

	// remove duplicate known existing dates
	seenKnownExistingDates := make(map[string]struct{})
	var outKnownExistingDates []JSONDate

	for _, date := range m.KnownExistingDates {
		if _, ok := seenKnownExistingDates[date.String()]; !ok {
			seenKnownExistingDates[date.String()] = struct{}{}
			outKnownExistingDates = append(outKnownExistingDates, date)
		}
	}

	m.KnownExistingDates = outKnownExistingDates

	// sort known existing dates
	sort.Slice(m.KnownExistingDates, func(i, j int) bool {
		return m.KnownExistingDates[i].String() < m.KnownExistingDates[j].String()
	})

	// sort missing dates
	sort.Slice(m.KnownMissingDates, func(i, j int) bool {
		return m.KnownMissingDates[i].Date.String() < m.KnownMissingDates[j].Date.String()
	})

	// remove duplicate rules
	seenRules := make(map[string]struct{})
	var outRules []Rule

	// new rules are added to the end of the slice and we want to keep the oldest ones
	// so we reverse the slice before iterating over it
	slices.Reverse(m.Rules)

	for _, rule := range m.Rules {
		key := fmt.Sprintf("%s-%s", rule.Date.String(), rule.Format)

		if _, ok := seenRules[key]; !ok {
			seenRules[key] = struct{}{}
			outRules = append(outRules, rule)
		}
	}

	m.Rules = outRules

	// sort rules
	sort.Slice(m.Rules, func(i, j int) bool {
		keyI := fmt.Sprintf("%s_%s", m.Rules[i].Date.String(), m.Rules[i].Format)
		keyJ := fmt.Sprintf("%s_%s", m.Rules[j].Date.String(), m.Rules[j].Format)

		return keyI < keyJ
	})
}
