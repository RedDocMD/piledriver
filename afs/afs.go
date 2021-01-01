package afs

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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
	name       string // Just of this directory/node
	parent     string // The rest of its path, before this node
	isDir      bool
	isTerminal bool
	driveID    string // ID corresponding to file in Google Drive
	children   map[string]*Node
}

// Tree represents the entire tree starting from a directory
type Tree struct {
	name string
	root *Node
}

func newNode(name, parent string, isDir, isTerminal bool) *Node {
	return &Node{
		name:       name,
		parent:     parent,
		isDir:      isDir,
		isTerminal: isTerminal,
		children:   make(map[string]*Node),
	}
}

// Splits paths into its directories
func splitPath(path string) []string {
	var partitions []string
	if runtime.GOOS == "windows" {
		driveRegex := regexp.MustCompile(`[A-Z]:\\`)
		if drive := driveRegex.FindString(path); drive != "" {
			partitions = append(partitions, drive[:2])
			path = path[4:]
		}

	}
	parts := strings.Split(path, string(filepath.Separator))
	if parts[len(parts)-1] == "" {
		partitions = append(partitions, parts[:len(parts)-1]...)
	} else {
		partitions = append(partitions, parts...)
	}
	return partitions
}

// Reverse of splitPath
func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	partsClone := make([]string, len(parts))
	copy(partsClone, parts)
	driveRegex := regexp.MustCompile(`[A-Z]:`)
	if drive := driveRegex.FindString(partsClone[0]); drive != "" {
		partsClone[0] = partsClone[0] + "\\"
	}
	return strings.Join(partsClone, string(filepath.Separator))
}

// NewTree creates a new tree from a given directory
// func NewTree(dir string) *Tree {
// }
