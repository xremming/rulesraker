package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func parseRun(cmd *cobra.Command, args []string) error {
	rules, err := openAndParseRules(cmd)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(dataDir, "MagicCompRules.json"))
	if err != nil {
		return err
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	err = enc.Encode(rules)
	if err != nil {
		return err
	}

	return nil
}

var parseCmd = &cobra.Command{
	Use:     "parse",
	Aliases: []string{"p"},
	Short:   "Parse the .txt rule file into a .json file",
	Args:    cobra.NoArgs,
	RunE:    parseRun,
}

func init() {
	rootCmd.AddCommand(parseCmd)
}
