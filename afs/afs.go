package afs

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
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
	name        string // Just of this directory/node
	parentPath  string // The rest of its path, before this node
	isDir       bool
	isRecursive bool   // Relevant only for directories
	driveID     string // ID corresponding to file in Google Drive
	children    map[string]*Node
	parentNode  *Node
}

// Tree represents the entire tree starting from a directory
type Tree struct {
	name string
	root *Node
}

func newNode(name, parentPath string, isDir, isRecursive bool, parentPtr *Node) *Node {
	return &Node{
		name:        name,
		parentPath:  parentPath,
		isDir:       isDir,
		isRecursive: isRecursive,
		children:    make(map[string]*Node),
		parentNode:  parentPtr,
		driveID:     "",
	}
}

func (node *Node) fullPath() string {
	return path.Join(node.parentPath, node.name)
}

func (node *Node) String() string {
	var b strings.Builder
	fmt.Fprint(&b, node.name)
	if node.driveID != "" {
		fmt.Fprintf(&b, " (%s0)", node.driveID)
	}
	if node.isDir {
		if node.isRecursive {
			fmt.Fprint(&b, " dr")
		} else {
			fmt.Fprint(&b, " d")
		}
	}
	return b.String()
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
			if stat, err := os.Stat(newPath); !os.IsNotExist(err) {
				newIsDir := stat.IsDir()
				newNode := newNode(name, currPath, newIsDir, node.isRecursive, node)
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

	rootNode := newNode(dirName, parent, true, isRecursive, nil)
	rootNode.extendNode()

	return &Tree{
		name: parent,
		root: rootNode,
	}
}

func (tree *Tree) String() string {
	prefixString := func(level uint) string {
		var builder strings.Builder
		for i := level; i > uint(0); i-- {
			fmt.Fprint(&builder, "|  ")
		}
		fmt.Fprint(&builder, "|--")
		return builder.String()
	}

	var dfs func(*Node, uint, *strings.Builder)
	dfs = func(node *Node, level uint, builder *strings.Builder) {
		pref := prefixString(level)
		fmt.Fprintf(builder, "%s%s\n", pref, node)
		for _, val := range node.children {
			dfs(val, level+1, builder)
		}
	}

	var b strings.Builder
	fmt.Println(&b, tree.name)
	dfs(tree.root, 0, &b)

	return b.String()
}

// AddPath adds a path to tree if the path is compatible with the tree
// path MUST be an absolute path, as is the assumption with all paths in the tree
// Returns true if the path was actually added
func (tree *Tree) AddPath(path string, isDir bool) bool {
	topPath := tree.root.fullPath()
	if !strings.HasPrefix(path, topPath) {
		return false
	}
	topPathParts := splitPath(topPath, string(filepath.Separator))
	pathParts := splitPath(path, string(filepath.Separator))

	truncatedParts := pathParts[len(topPathParts):]

	var addPath func(node *Node, parts []string)
	addPath = func(node *Node, parts []string) {
		if len(parts) == 0 {
			return
		}
		var thisPartIsDir bool
		if len(parts) == 1 {
			thisPartIsDir = isDir
		} else {
			thisPartIsDir = true
		}
		childNode := newNode(parts[0], node.fullPath(), thisPartIsDir, node.isRecursive, node)
		node.children[parts[0]] = childNode
		addPath(childNode, parts[1:])
		childNode.extendNode()
	}

	var findNode func(node *Node, parts []string) (*Node, []string)
	findNode = func(node *Node, parts []string) (*Node, []string) {
		if len(parts) == 0 {
			return nil, nil
		}
		first := parts[0]
		child, ok := node.children[first]
		if !ok {
			return node, parts
		}
		return findNode(child, parts[1:])
	}

	addNode, remaining := findNode(tree.root, truncatedParts)
	if addNode == nil {
		return false
	}
	addPath(addNode, remaining)
	return true
}

// Given a path, searches if it is in the tree
func (tree *Tree) findPath(path string) (*Node, bool) {
	topPath := tree.root.fullPath()
	if !strings.HasPrefix(path, topPath) {
		return nil, false
	}
	topPathParts := splitPath(topPath, string(filepath.Separator))
	pathParts := splitPath(path, string(filepath.Separator))

	truncatedParts := pathParts[len(topPathParts):]

	var findPathInternal func(node *Node, parts []string) (*Node, bool)
	findPathInternal = func(node *Node, parts []string) (*Node, bool) {
		if len(parts) == 0 {
			return node, true
		}
		next := parts[0]
		child, ok := node.children[next]
		if !ok {
			return nil, false
		}
		return findPathInternal(child, parts[1:])
	}

	return findPathInternal(tree.root, truncatedParts)
}
