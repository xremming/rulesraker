package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dataDir string
)

var rootCmd = &cobra.Command{
	Use: "rulesraker",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return os.MkdirAll(dataDir, 0o700)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dataDir, "data", "d", "data",
		"directory where the comprehensive rules are stored",
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
