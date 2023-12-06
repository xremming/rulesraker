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
)

func newlineToBR(s string) template.HTML {
	return template.HTML(strings.ReplaceAll(s, "\n", "<br>"))
}

func formatTime(layout string, t time.Time) string {
	return t.Format(layout)
}

var (
	outputDir   string
	templateDir string
	publicDir   string
)

func buildRun(cmd *cobra.Command, args []string) error {
	cmd.Println("updating symbols from Scryfall")
	symbolReplacer, err := getSymbolsReplacer()
	if err != nil {
		return err
	}

	rules, err := openAndParseRules(cmd)
	if err != nil {
		return err
	}

	cmd.Println("rendering index.html")
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

	indexHTML, err := os.Create(filepath.Join(outputDir, "index.html"))
	if err != nil {
		return err
	}
	defer indexHTML.Close()

	err = tmpl.ExecuteTemplate(indexHTML, "index.html", map[string]any{
		"Title":         "Rulesraker - Magic: the Gathering Comprehensive Rules",
		"Description":   "A fast and easy to use interface for Magic: the Gathering's Comprehensive Rules.",
		"RulesURL":      rulesURL,
		"EffectiveDate": rules.EffectiveDate,
		"Rules":         rules.Rules,
		"Glossary":      rules.Glossary,
		"Credits":       rules.Credits,
	})
	if err != nil {
		return err
	}

	// copy all files from publicDir to outputDir
	public := os.DirFS(publicDir)
	err = fs.WalkDir(public, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if path == "index.html" {
			cmd.Println("skipping index.html from the public directory")
			return nil
		}

		cmd.Printf("copying %q from public directory to the output directory", path)
		inp, err := public.Open(path)
		if err != nil {
			return err
		}
		defer inp.Close()

		out, err := os.Create(filepath.Join(outputDir, path))
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
	if err != nil {
		return err
	}

	return nil
}

var buildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Short:   "Build the site from the .txt file",
	Args:    cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return os.MkdirAll(outputDir, 0o755)
	},
	RunE: buildRun,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&outputDir, "out", "o", "dist",
		"directory where the site will be rendered",
	)
	buildCmd.Flags().StringVar(&templateDir, "template", "template",
		"directory which contains the templates for rendering",
	)
	buildCmd.Flags().StringVar(&publicDir, "public", "public",
		"directory which contains files that will be copied as-is to the output directory",
	)
}
