package backup

import (
	"github.com/RedDocMD/piledriver/afs"
	"github.com/RedDocMD/piledriver/utils"
	"google.golang.org/api/drive/v3"
)

func BackupToDrive(localTree, driveTree *afs.Tree, service *drive.Service, rootID string) error {
	if driveTree == nil {
		return backupNode(localTree.Root(), service, localTree.RootName(), rootID)
	}
	return nil
}

func backupNode(node *afs.Node, service *drive.Service, localPath, parentID string) error {
	if node.IsDir() {
		id, err := utils.CreateFolder(service, localPath, parentID)
		if err != nil {
			return err
		}
		children := node.Children()
		for name := range children {
			childNode := children[name]
			localPathParts := afs.SplitPathPlatform(localPath)
			newPath := afs.JoinPathPlatform(append(localPathParts, childNode.Name()), true)
			if err := backupNode(childNode, service, newPath, id); err != nil {
				return err
			}
		}
	}
	return nil
}
