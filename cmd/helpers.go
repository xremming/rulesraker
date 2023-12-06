package cmd

import (
	"os"
	"path/filepath"

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
