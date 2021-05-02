package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Piledriver",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Piledriver 0.1.0")
	},
	DisableFlagsInUseLine: true,
}
