package main

import (
	"log"
	"os"
	"path"

	"github.com/RedDocMD/piledriver/config"
	"github.com/RedDocMD/piledriver/utils"
	"github.com/spf13/viper"
)

func main() {
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

	var config config.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Error in config file: %s\n", err)
	}

	state := utils.NewState()
	state.InitService(config.TokenPath)
	state.InitWatcher()
	for _, dir := range config.Directories {
		state.AddDir(dir.Local, dir.Recursive)
	}
	go utils.ExecuteEvents(state.FileEvents)
	utils.WatchLoop(state)
}
