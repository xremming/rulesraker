package cmd

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	watch       bool
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

	build := func() error {
		indexHTML, err := os.Create(filepath.Join(outputDir, "index.html"))
		if err != nil {
			return err
		}
		defer indexHTML.Close()

		cmd.Println("rendering index.html")
		err = renderIndex(indexHTML, rules, symbolReplacer)
		if err != nil {
			return err
		}

		// copy all files from publicDir to outputDir
		public := os.DirFS(publicDir)
		err = copyRecursive(cmd, public, outputDir)
		if err != nil {
			return err
		}

		return nil
	}

	err = build()
	if err != nil {
		cmd.PrintErrf("error when building: %v\n", err)
		if !watch {
			return err
		}
	}

	if watch {
		cmd.Println("watching for file system changes")

		w, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		err = w.Add(publicDir)
		if err != nil {
			return err
		}

		err = w.Add(templateDir)
		if err != nil {
			return err
		}

		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return nil
				}

				cmd.Printf("file %q changed, rebuilding\n", event.Name)

				err = build()
				if err != nil {
					cmd.PrintErrf("error when building: %v\n", err)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return nil
				}

				return err
			}
		}
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
	buildCmd.Flags().BoolVarP(&watch, "watch", "w", false,
		"watch for file system changes",
	)
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
