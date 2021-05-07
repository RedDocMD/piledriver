package cmd

import (
	"log"

	"github.com/RedDocMD/piledriver/config"
	"github.com/RedDocMD/piledriver/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate Piledriver to use your Google Drive",
	Long: `This command authenticates Piledriver via OAuth 2.0. This way
it has access to your Google Drive, but limited to only the files
Piledriver has created. So Piledriver cannot see files that it has
not synced`,
	Run: func(cmd *cobra.Command, args []string) {
		var config config.Config
		err := viper.Unmarshal(&config)
		if err != nil {
			log.Fatalf("Unable to parse config: %s\n", err)
		}
		utils.AuthorizeApp(config.TokenPath)
	},
	DisableFlagsInUseLine: true,
}
