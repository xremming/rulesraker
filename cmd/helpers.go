package cmd

import (
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

func newlineToBR(s string) template.HTML {
	return template.HTML(strings.ReplaceAll(s, "\n", "<br>"))
}

func formatTime(layout string, t time.Time) string {
	return t.Format(layout)
}

func renderIndex(w io.Writer, rules parser.Rules, symbolReplacer *strings.Replacer) error {
	tmpl, err := template.New("").
		Funcs(template.FuncMap{
			"formatTime":  formatTime,
			"newlineToBR": newlineToBR,
			"replaceSymbols": func(s string) template.HTML {
				return template.HTML(symbolReplacer.Replace(s))
			},
		}).
		ParseFS(os.DirFS(templateDir), "*.html")
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, "index.html", map[string]any{
		"Title":         "Rulesraker - Magic: the Gathering Comprehensive Rules",
		"Description":   "A fast and easy interface to Magic: the Gathering's Comprehensive Rules.",
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
