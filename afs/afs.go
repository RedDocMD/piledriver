package afs

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

// NewTree creates a new tree from a given directory
// func NewTree(dir string) *Tree {
// }
