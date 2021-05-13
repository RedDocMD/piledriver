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
		state.InitWatcher()
		for _, dir := range config.Directories {
			state.AddDir(dir.Local)
		}

		// Run the watch loop to accumulate changes in the init period
		go utils.WatchLoop(state)

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

		driveFiles, err := utils.QueryAllContents(state.Service())
		if err != nil {
			log.Fatalf("Failed to retrieve file list from Drive: %s\n", err)
		}
		log.Println("Retrieved file info from Drive")

		type TreeName struct {
			tree       *afs.Tree
			remoteName string
		}

		driveTreesNames := make(map[string]TreeName)
		for _, dir := range config.Directories {
			tree, err := afs.NewTreeFromDrive(driveFiles, dir.Remote)
			if err != nil {
				log.Println(err)
				driveTreesNames[dir.Local] = TreeName{nil, dir.Remote}
			} else {
				driveTreesNames[dir.Local] = TreeName{tree, dir.Remote}
			}
		}

		// First make sure that the local and drive trees have the same structure
		updated := false
		for dir := range driveTreesNames {
			driveTreeName := driveTreesNames[dir]
			localTree, _ := state.Tree(dir)
			if driveTreeName.tree == nil || !localTree.EqualsIgnore(driveTreeName.tree, true) {
				updated = true
				log.Printf("Backing up tree in %s ...\n", localTree.RootPath())
				err = backup.ToDrive(
					localTree,
					driveTreeName.tree,
					driveTreeName.remoteName,
					state.Service(),
					rootFolderID,
				)
				if err != nil {
					log.Fatalf("Failed to perform force backup: %s", err)
				}
				log.Printf("Backed up tree in %s\n", localTree.RootPath())
			}
		}

		// Update the drive trees to reflect the changes
		if updated {
			driveFiles, err = utils.QueryAllContents(state.Service())
			if err != nil {
				log.Fatalf("Failed to retrieve file list from Drive: %s\n", err)
			}
			log.Println("Retrieved file info from Drive")
			for _, dir := range config.Directories {
				tree, err := afs.NewTreeFromDrive(driveFiles, dir.Remote)
				if err != nil {
					log.Fatalf("Failed to find drive tree rooted at %s corresponding to local tree at %s\n", dir.Remote, dir.Local)
				}
				driveTreesNames[dir.Local] = TreeName{tree, dir.Remote}
			}
		}

		// Attach the drive ID's to the local tree
		for name := range driveTreesNames {
			localTree, _ := state.Tree(name)
			driveTree := driveTreesNames[name].tree
			backup.AttachIDS(localTree, driveTree)
			log.Printf("Attached ID's to tree with root path %s\n", localTree.RootPath())
		}

		// Check if the local version of files is more recent than the drive version
		for name := range driveTreesNames {
			localTree, _ := state.Tree(name)
			driveTree := driveTreesNames[name].tree
			err := localTree.CalculateChecksums()
			if err != nil {
				log.Fatalf("Failed to calculate local tree checksums: %s\n", err)
			}
			log.Printf("Calculated checksums for tree rooted at %s\n", localTree.RootPath())
			err = backup.UpdateDriveTree(localTree, driveTree, state.Service())
			if err != nil {
				log.Fatalf("Failed to update changed files for tree rooted at %s: %s\n", localTree.RootPath(), err)
			}
			log.Printf("Updated to drive, tree rooted at %s\n", localTree.RootPath())
		}

		// const noOfWorkers int = 12
		// for i := 0; i < noOfWorkers; i++ {
		// 	go utils.ExecuteEvents(state)
		// }
		utils.ExecuteEvents(state)

		// Now just keep on running
		// var wg sync.WaitGroup
		// wg.Add(1)
		// wg.Wait()
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
