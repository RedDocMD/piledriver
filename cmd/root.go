package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/RedDocMD/piledriver/afs"
	"github.com/RedDocMD/piledriver/backup"
	"github.com/RedDocMD/piledriver/config"
	"github.com/RedDocMD/piledriver/utils"
	"github.com/denisbrodbeck/machineid"
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

		driveFiles, err := utils.QueryAllContents(state.Service())
		if err != nil {
			log.Fatalf("Failed to retrieve file list from Drive: %s\n", err)
		}
		log.Println("Retrieved file info from Drive")

		rootFolder := fmt.Sprintf("piledriver-%s", config.MachineIdentifier)
		rootFolderID, err := utils.QueryFileID(state.Service(), rootFolder)
		if err != nil && err.Error() == fmt.Sprintf("Didn't find %s in you Drive", rootFolder) {
			rootFolderID, err = utils.CreateFolder(state.Service(), rootFolder)
			if err != nil {
				log.Fatalf("Failed to create rootFolder %s: %s\n", rootFolder, err)
			}
			log.Printf("Created %s as rootFolder\n", rootFolder)
		} else if err != nil {
			log.Fatalf("Failed to query rootFolder %s: %s\n", rootFolder, err)
		}

		state.InitWatcher()
		driveTrees := make(map[string]*afs.Tree)
		for _, dir := range config.Directories {
			state.AddDir(dir.Local)
			tree, err := afs.NewTreeFromDrive(driveFiles, dir.Remote)
			if err != nil {
				log.Println(err)
				driveTrees[dir.Local] = nil
			} else {
				driveTrees[dir.Local] = tree
			}
		}

		// First make sure that the local and drive trees have the same structure
		for dir := range driveTrees {
			driveTree := driveTrees[dir]
			localTree, _ := state.Tree(dir)
			if driveTree == nil || !localTree.Equals(driveTree) {
				err = backup.BackupToDrive(localTree, driveTree, state.Service(), rootFolderID)
				log.Printf("Backeing up tree in %s ...\n", localTree.RootPath())
				if err != nil {
					log.Fatalf("Failed to perform force backup: %s", err)
				}
				log.Printf("Backed up tree in %s\n", localTree.RootPath())
			}
		}
		// Update the drive trees to reflect the changes
		// Attach the drive ID's to the local tree
		// Check if the local version of files is more recent than the drive version

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
	const randomString string = "XcK2YkF8rkyCQRlX9qn9"
	machineID, err := machineid.ProtectedID(randomString)
	if err != nil {
		log.Println("Failed to retrieve machine ID")
	}
	viper.SetDefault("machineIdentifier", machineID)
}
