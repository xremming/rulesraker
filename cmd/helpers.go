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
		fmt.Sprintf("script-src 'self' 'nonce-%s'", nonce),
		"img-src 'self' https://svgs.scryfall.io https://cards.scryfall.io",
		"connect-src 'self' https://api.scryfall.com",
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

func lower(s any) string {
	v := reflect.ValueOf(s)
	return strings.ToLower(v.String())
}

func renderIndex(w io.Writer, rules parser.Rules, symbolReplacer *strings.Replacer) error {
	tmpl, err := template.New("").
		Funcs(template.FuncMap{
			"formatTime":  formatTime,
			"newlineToBR": newlineToBR,
			"startsWith":  startsWith,
			"lower":       lower,
			"replaceSymbols": func(s string) template.HTML {
				return template.HTML(symbolReplacer.Replace(s))
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
	return fs.WalkDir(from, ".", func(path string, d fs.DirEntry, err error) error {
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
