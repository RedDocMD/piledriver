package backup

import (
	"github.com/RedDocMD/piledriver/afs"
	"github.com/RedDocMD/piledriver/utils"
	"google.golang.org/api/drive/v3"
)

// ToDrive is a sort of force backup, where the local tree and the drive tree is made consistent.
// This is only necessary:
//   - When a new folder is added to be backuped (or for the first time Piledriver is run)
//   - When you make changes in the local fs with Piledriver off
//   - When you manually edit the files in Drive
func ToDrive(
	localTree, driveTree *afs.Tree,
	remoteRootName string,
	service *drive.Service,
	rootID string) error {

	rootPath := localTree.RootPath()
	if driveTree == nil {
		return backupNode(
			localTree.Root(),
			service,
			rootPath,
			remoteRootName,
			rootID,
			true,
		)
	}
	pathParts := afs.SplitPathPlatform(rootPath)
	var backupOnMismatch func(localNode, driveNode *afs.Node) error
	backupOnMismatch = func(localNode, driveNode *afs.Node) error {
		if localNode != localTree.Root() && localNode.Name() != driveNode.Name() {
			return backupNode(
				localNode,
				service,
				afs.JoinPathPlatform(pathParts, true),
				remoteRootName,
				driveNode.Parent().DriveID(),
				false,
			)
		}
		localChildren := localNode.Children()
		driveChildren := driveNode.Children()
		var driveChildrenCovered []*afs.Node
		for localName := range localChildren {
			localChild := localChildren[localName]
			pathParts = append(pathParts, localChild.Name())
			var err error
			if driveChild, ok := driveChildren[localName]; !ok {
				err = backupNode(
					localChild,
					service,
					afs.JoinPathPlatform(pathParts, true),
					remoteRootName,
					driveNode.DriveID(),
					false,
				)
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
				err := utils.DeleteFileOrFolder(service, driveChild.DriveID())
				if err != nil {
					return err
				}
			}
		}
		return nil
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

func backupNode(
	node *afs.Node,
	service *drive.Service,
	localPath, rootRemoteName, parentID string,
	isRoot bool) error {

	if node.IsDir() {
		var id string
		var err error
		if isRoot {
			id, err = utils.CreateFolder(service, rootRemoteName, parentID)
		} else {
			id, err = utils.CreateFolder(service, localPath, parentID)
		}
		if err != nil {
			return err
		}
		children := node.Children()
		for name := range children {
			childNode := children[name]
			localPathParts := afs.SplitPathPlatform(localPath)
			newPath := afs.JoinPathPlatform(append(localPathParts, childNode.Name()), true)
			err := backupNode(
				childNode,
				service,
				newPath,
				rootRemoteName,
				id,
				false,
			)
			if err != nil {
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

// AttachIDS attaches all the ids from the Drive AFS to the local AFS.
// It assumes that the two trees have the same structure, ie, they return
// true for drive.EqualsIgnore(local, true).
func AttachIDS(localTree, driveTree *afs.Tree) {
	var attach func(localNode, driveNode *afs.Node)
	attach = func(localNode, driveNode *afs.Node) {
		localNode.SetDriveID(driveNode.DriveID())
		localChildren := localNode.Children()
		driveChildren := driveNode.Children()
		for childName := range localChildren {
			localChild := localChildren[childName]
			driveChild := driveChildren[childName]
			attach(localChild, driveChild)
		}
	}
	attach(localTree.Root(), driveTree.Root())
}

// UpdateDriveTree updates the drive tree to match the local tree,
// updating files when they mismatch (from the checksums).
// It assumes that the two trees have the same structure, ie, they return
// true for drive.EqualsIgnore(local, true).
// It also assumes that the localTree has the driveID's in place
func UpdateDriveTree(localTree, driveTree *afs.Tree, service *drive.Service) error {
	var update func(localNode, driveNode *afs.Node) error
	pathParts := afs.SplitPathPlatform(localTree.RootPath())
	pathParts = pathParts[0 : len(pathParts)-1]
	update = func(localNode, driveNode *afs.Node) error {
		pathParts = append(pathParts, localNode.Name())
		if localNode.IsDir() {
			localChildren := localNode.Children()
			driveChildren := driveNode.Children()
			for childName := range localChildren {
				localChild := localChildren[childName]
				driveChild := driveChildren[childName]
				err := update(localChild, driveChild)
				if err != nil {
					return err
				}
			}
		} else {
			if localNode.Checksum() != driveNode.Checksum() {
				file, err := utils.UpdateFile(
					service,
					afs.JoinPathPlatform(pathParts, true),
					localNode.DriveID(),
				)
				if err != nil {
					return err
				}
				newChecksum := file.AppProperties["md5sum"]
				driveNode.SetChecksum(newChecksum)
				localNode.SetChecksum(newChecksum)
			}
		}
		pathParts = pathParts[0 : len(pathParts)-1]
		return nil
	}
	return update(localTree.Root(), driveTree.Root())
}
