package afs

import (
	"log"
	"os"
	"path"
	"path/filepath"
)

// An Abstract File System which mimics a file system tree
// but is created to store information about paths and their corresponding
// Google Drive ID.
// An internal node in the AFS is a directory which is to be watched recursively.
// A leaf node is either a file or a directory that will not be watched (As
// a consequence, its parent must be a directory that must be non-recursively watched)

// Node is a single node in the AFS, corresponding to a unique path in the OS,
// and consequently in Google Drive
type Node struct {
	name        string // Just of this directory/node
	parent      string // The rest of its path, before this node
	isDir       bool
	isRecursive bool   // Relevant only for directories
	driveID     string // ID corresponding to file in Google Drive
	children    map[string]*Node
}

// Tree represents the entire tree starting from a directory
type Tree struct {
	name string
	root *Node
}

func newNode(name, parent string, isDir, isRecursive bool) *Node {
	return &Node{
		name:        name,
		parent:      parent,
		isDir:       isDir,
		isRecursive: isRecursive,
		children:    make(map[string]*Node),
		driveID:     "",
	}
}

func (node *Node) fullPath() string {
	return path.Join(node.parent, node.name)
}

// Explores a node, to extend it downwards
// Nothing to explore if it's a file
// If it's a directory and is non-recursive, add all its files (but not directories) as children
// It it's a directory and is recursive, add all its files and recursively explore its sub-directories
func (node *Node) extendNode() {
	if node.isDir {
		currPath := node.fullPath()
		dir, err := os.Open(currPath)
		defer dir.Close()
		if err != nil {
			log.Fatal(err)
		}
		contents, err := dir.Readdirnames(-1)
		if err != nil {
			log.Fatal(err)
		}
		for _, name := range contents {
			newPath := path.Join(currPath, name)
			if stat, err := os.Stat(newPath); os.IsExist(err) {
				newIsDir := stat.IsDir()
				newNode := newNode(name, currPath, newIsDir, node.isRecursive)
				node.children[name] = newNode
				if node.isRecursive {
					newNode.extendNode()
				}
			}
		}
	}
}

// NewTree creates a new tree from a given directory
func NewTree(dir string, isRecursive bool) *Tree {
	parts := splitPath(dir, string(filepath.Separator))
	parent := joinPath(parts[:len(parts)-1], string(filepath.Separator), true)
	dirName := parts[len(parts)-1]

	rootNode := newNode(parent, dirName, true, isRecursive)
	rootNode.extendNode()

	return &Tree{
		name: parent,
		root: rootNode,
	}
}
