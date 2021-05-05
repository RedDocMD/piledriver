package cmd

import (
	"log"
	"os"
	"path"

	"github.com/RedDocMD/piledriver/config"
	"github.com/RedDocMD/piledriver/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:                   "piledriver",
	Short:                 "Piledriver is a Google Drive sync-daemon",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		var config config.Config
		err := viper.Unmarshal(&config)
		if err != nil {
			log.Fatalf("Error in config file: %s\n", err)
		}

		state := utils.NewState()
		state.InitService(config.TokenPath)
		state.InitWatcher()
		for _, dir := range config.Directories {
			state.AddDir(dir.Local)
		}

		const noOfWorkers int = 12
		for i := 0; i < noOfWorkers; i++ {
			go utils.ExecuteEvents(state.FileEvents)
		}
		utils.WatchLoop(state)
	},
}

// Execute is the top-level command execute - call this from main
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(authCmd)
}

func initConfig() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to find user home directory: %s\n", err)
	}

	viper.SetConfigType("json")
	viper.SetConfigName("piledriver")
	viper.AddConfigPath(homedir)
	viper.AddConfigPath(path.Join(homedir, ".config"))
	viper.AddConfigPath(path.Join(homedir, ".config", "piledriver"))

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %s\n", err)
	}

	viper.SetDefault("tokenPath", path.Join(homedir, ".piledriver.token"))
}
