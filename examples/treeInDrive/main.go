package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/RedDocMD/piledriver/afs"
	"github.com/RedDocMD/piledriver/utils"
)

func main() {
	searchDir := afs.JoinPathPlatform([]string{"examples", "treeInDrive", "tree_dir"}, false)
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Failed to find user home dir:", err)
	}
	homedirParts := afs.SplitPathPlatform(homedir)
	tokenPath := afs.JoinPathPlatform(append(homedirParts, []string{".config", ".piledriver.token"}...), true)
	service := utils.GetDriveService(tokenPath)

	parentID := make(map[string]string)

	filepath.WalkDir(searchDir, func(path string, d fs.DirEntry, err error) error {
		parts := afs.SplitPathPlatform(path)
		parentPath := afs.JoinPathPlatform(parts[0:len(parts)-1], false)
		var id string
		if d.IsDir() {
			if parent, ok := parentID[parentPath]; ok {
				id, err = utils.CreateFolder(service, path, parent)
			} else {
				id, err = utils.CreateFolder(service, path)
			}
		} else {
			parent := parentID[parentPath]
			id, err = utils.CreateFile(service, path, parent)
		}
		parentID[path] = id
		if err != nil {
			log.Fatalf("Failed to make %s: %s", path, err)
		}
		log.Printf("Created %s with ID %s", path, id)
		return nil
	})
}
