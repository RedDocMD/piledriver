package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                   "piledriver",
	Short:                 "Piledriver is a Google Drive sync-daemon",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
