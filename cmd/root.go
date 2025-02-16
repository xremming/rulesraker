package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	dataDir    string
	archiveDir string
)

var rootCmd = &cobra.Command{
	Use: "rulesraker",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return errors.Join(
			os.MkdirAll(dataDir, 0o755),
			os.MkdirAll(archiveDir, 0o755),
			os.MkdirAll(filepath.Join(archiveDir, "docx"), 0o755),
			os.MkdirAll(filepath.Join(archiveDir, "pdf"), 0o755),
			os.MkdirAll(filepath.Join(archiveDir, "txt"), 0o755),
		)
	},
}

func init() {
	var err error

	rootCmd.PersistentFlags().StringVarP(&dataDir, "data-dir", "d", "data",
		"directory where the comprehensive rules are stored",
	)
	err = rootCmd.MarkFlagDirname("data-dir")
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringVarP(&archiveDir, "archive-dir", "a", "archive",
		"directory to store stores the archived rules",
	)
	err = rootCmd.MarkFlagDirname("archive-dir")
	if err != nil {
		panic(err)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
