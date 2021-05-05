package backup

import (
	"github.com/RedDocMD/piledriver/afs"
	"github.com/RedDocMD/piledriver/utils"
	"google.golang.org/api/drive/v3"
)

// BackupToDrive is a sort of force backup, where the local tree and the drive tree is made consistent.
// This is only necessary:
//   - When a new folder is added to be backuped (or for the first time Piledriver is run)
//   - When you make changes in the local fs with Piledriver off
//   - When you manually edit the files in Drive
func BackupToDrive(localTree, driveTree *afs.Tree, service *drive.Service, rootID string) error {
	rootPath := localTree.RootPath()
	if driveTree == nil {
		return backupNode(localTree.Root(), service, rootPath, rootID)
	}
	pathParts := afs.SplitPathPlatform(rootPath)
	var backupOnMismatch func(localNode, driveNode *afs.Node) error
	backupOnMismatch = func(localNode, driveNode *afs.Node) error {
		if localNode.Name() != driveNode.Name() {
			err := backupNode(localNode, service, afs.JoinPathPlatform(pathParts, true), driveNode.Parent().DriveID())
			return err
		} else {
			localChildren := localNode.Children()
			driveChildren := driveNode.Children()
			var driveChildrenCovered []*afs.Node
			for localName := range localChildren {
				localChild := localChildren[localName]
				pathParts = append(pathParts, localChild.Name())
				var err error
				if driveChild, ok := driveChildren[localName]; !ok {
					err = backupNode(localChild, service, afs.JoinPathPlatform(pathParts, true), driveNode.DriveID())
				} else {
					driveChildrenCovered = append(driveChildrenCovered, driveChild)
					err = backupOnMismatch(localChild, driveChild)
				}
				if err != nil {
					return err
				}
				pathParts = pathParts[0 : len(pathParts)-1]
			}
			// Now remove extra nodes from Drive Tree
			for driveName := range driveChildren {
				driveChild := driveChildren[driveName]
				if !nodeIsPresent(driveChildrenCovered, driveChild) {
					utils.DeleteFileOrFolder(service, driveChild.DriveID())
				}
			}
			return nil
		}
	}
	return backupOnMismatch(localTree.Root(), driveTree.Root())
}

func nodeIsPresent(list []*afs.Node, node *afs.Node) bool {
	for _, listNode := range list {
		if listNode == node {
			return true
		}
	}
	return false
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
	} else {
		if _, err := utils.CreateFile(service, localPath, parentID); err != nil {
			return err
		}
	}
	return nil
}
