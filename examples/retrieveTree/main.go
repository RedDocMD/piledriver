package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RedDocMD/piledriver/afs"
	"github.com/RedDocMD/piledriver/utils"
)

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Failed to find user home dir:", err)
	}
	homedirParts := afs.SplitPathPlatform(homedir)
	tokenPath := afs.JoinPathPlatform(append(homedirParts, []string{".config", ".piledriver.token"}...), true)
	service := utils.GetDriveService(tokenPath)

	files, err := utils.QueryAllContents(service)
	if err != nil {
		log.Fatalln("Failed to retrieve file list:", err)
	}
	for _, file := range files {
		fmt.Printf("%s => %s (parent = %s)\n", file.Name, file.Id, file.Parents[0])
	}
}
