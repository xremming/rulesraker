package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xremming/rulesraker/parser"
)

func openAndParseRules(cmd *cobra.Command) (parser.Rules, error) {
	cmd.Println("opening rules text")
	fp, err := os.Open(filepath.Join(dataDir, "MagicCompRules.txt"))
	if err != nil {
		return parser.Rules{}, err
	}
	defer fp.Close()

	cmd.Println("parsing rules")
	rules, err := parser.Parse(fp)
	if err != nil {
		return parser.Rules{}, err
	}

	return rules, nil
}

func makeCSP() (string, string) {
	bytes := make([]byte, 12)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	nonce := base64.URLEncoding.EncodeToString(bytes)

	cspLines := []string{
		"default-src 'self'",
		fmt.Sprintf("script-src 'self' 'nonce-%s' https://static.cloudflareinsights.com", nonce),
		"img-src 'self' https://svgs.scryfall.io https://cards.scryfall.io",
		"connect-src 'self' https://api.scryfall.com https://cloudflareinsights.com",
		"object-src 'none'",
		"child-src 'none'",
	}

	return nonce, strings.Join(cspLines, "; ") + ";"
}

func formatTime(layout string, t time.Time) string {
	return t.Format(layout)
}

func newlineToBR(s string) template.HTML {
	return template.HTML(strings.ReplaceAll(s, "\n", "<br>"))
}

func startsWith(prefix, s string) bool {
	return strings.HasPrefix(s, prefix)
}

func asString(s any) string {
	return reflect.ValueOf(s).String()
}

func lower(s any) string {
	return strings.ToLower(asString(s))
}

var linkifyRegexp = regexp.MustCompile(`(?i)([a-z\.]+\.com[a-z-\/]*)`)

func linkify(s any) template.HTML {
	return template.HTML(linkifyRegexp.ReplaceAllString(asString(s), `<a href="https://$1" target="_blank">$0</a>`))
}

var numberRegexp = regexp.MustCompile(`(\d{3})(\.((\d+)(\w+)?)(–(\d+|\w)+)?)?`)

func parseNumber(rule string) string {
	matches := numberRegexp.FindStringSubmatch(rule)

	major := matches[1]
	minor := matches[4]
	letter := matches[5]

	if minor == "" && letter == "" {
		return fmt.Sprintf("%v.", major)
	}

	if letter == "" {
		return fmt.Sprintf("%v.%v.", major, minor)
	}

	return fmt.Sprintf("%v.%v%v", major, minor, letter)
}

var sectionRefRegexp = regexp.MustCompile(`section (\d)`)

func ruleLinks(s any) template.HTML {
	text := numberRegexp.ReplaceAllStringFunc(asString(s), func(s string) string {
		return fmt.Sprintf(`<a href="#%s">%s</a>`, parseNumber(s), s)
	})

	return template.HTML(sectionRefRegexp.ReplaceAllString(text, `<a href="#$1.">$0</a>`))
}

func renderIndex(w io.Writer, rules parser.Rules, symbolReplacer *strings.Replacer) error {
	tmpl, err := template.New("").
		Funcs(template.FuncMap{
			"formatTime":  formatTime,
			"newlineToBR": newlineToBR,
			"startsWith":  startsWith,
			"lower":       lower,
			"linkify":     linkify,
			"ruleLinks":   ruleLinks,
			"replaceSymbols": func(s any) template.HTML {
				return template.HTML(symbolReplacer.Replace(asString(s)))
			},
		}).
		ParseFS(os.DirFS(templateDir), "*.html")
	if err != nil {
		return err
	}

	nonce, csp := makeCSP()

	return tmpl.ExecuteTemplate(w, "index.html", map[string]any{
		"Title":         "Rulesraker - Magic: the Gathering Comprehensive Rules",
		"Description":   "A fast and easy interface to Magic: the Gathering's Comprehensive Rules.",
		"CSP":           csp,
		"Nonce":         nonce,
		"RulesURL":      rulesURL,
		"EffectiveDate": rules.EffectiveDate,
		"Rules":         rules.Rules,
		"Glossary":      rules.Glossary,
		"Credits":       rules.Credits,
	})
}

func copyRecursive(cmd *cobra.Command, from fs.FS, to string) error {
	return fs.WalkDir(from, ".", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		if path == "index.html" {
			cmd.Println("skipping index.html from the public directory")
			return nil
		}

		cmd.Printf("copying %q from public directory to the output directory\n", path)
		inp, err := from.Open(path)
		if err != nil {
			return err
		}
		defer inp.Close()

		out, err := os.Create(filepath.Join(to, path))
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, inp)
		if err != nil {
			return err
		}

		return nil
	})
}
