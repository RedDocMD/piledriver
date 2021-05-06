package afs

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v3"
)

// An Abstract File System which mimics a file system tree
// but is created to store information about paths and their corresponding
// Google Drive ID.
// An internal node in the AFS is a directory which is to be watched recursively.
// A leaf node is a file

// Node is a single node in the AFS, corresponding to a unique path in the OS,
// and consequently in Google Drive
type Node struct {
	name       string // Just of this directory/node
	isDir      bool
	driveID    string // ID corresponding to file in Google Drive
	children   map[string]*Node
	parentNode *Node
}

// Tree represents the entire tree starting from a directory
type Tree struct {
	name string
	root *Node
}

func newNode(name string, isDir bool, parentPtr *Node) *Node {
	return &Node{
		name:       name,
		isDir:      isDir,
		children:   make(map[string]*Node),
		parentNode: parentPtr,
		driveID:    "",
	}
}

// IsDir returns whether this node is a directory.
func (node *Node) IsDir() bool {
	return node.isDir
}

// Name returns the name field for node
func (node *Node) Name() string {
	return node.name
}

// Parent returns the parentNode field for this node.
func (node *Node) Parent() *Node {
	return node.parentNode
}

// DriveID returns the dirveID for this node.
func (node *Node) DriveID() string {
	return node.driveID
}

// SetDriveID sets the driveID for this node
func (node *Node) SetDriveID(id string) {
	node.driveID = id
}

func (node *Node) String() string {
	var b strings.Builder
	fmt.Fprint(&b, node.name)
	if node.driveID != "" {
		fmt.Fprintf(&b, " (%s0)", node.driveID)
	}
	if node.isDir {
		fmt.Fprint(&b, " d")
	}
	return b.String()
}

func (node *Node) Children() map[string]*Node {
	return node.children
}

// NewTree creates a new tree from a given directory
func NewTree(dir string) *Tree {
	parts := SplitPathPlatform(dir)
	parent := JoinPathPlatform(parts[:len(parts)-1], true)
	dirName := parts[len(parts)-1]

	rootNode := newNode(dirName, true, nil)

	return &Tree{
		name: parent,
		root: rootNode,
	}
}

// TreeFromDrive reconstructs the tree from the list of files
// retrieved from Google Drive
func NewTreeFromDrive(files []*drive.File, rootPath string) (*Tree, error) {
	rootId := ""
	rootPathParts := SplitPathPlatform(rootPath)
	rootName := rootPathParts[len(rootPathParts)-1]
	childrenOf := make(map[string][]*drive.File)
	for _, file := range files {
		if rootName == file.Name {
			rootId = file.Id
		}
		parentId := file.Parents[0]
		childrenOf[parentId] = append(childrenOf[parentId], file)
	}
	if rootId == "" {
		return nil, fmt.Errorf("can't find id for %s", rootPath)
	}

	rootNode := newNode(rootName, true, nil)
	rootNode.driveID = rootId

	// Do BFS
	queue := []*Node{rootNode}
	for len(queue) != 0 {
		node := queue[0]
		queue = queue[1:]
		children, ok := childrenOf[node.driveID]
		if ok {
			for _, child := range children {
				isDir := child.MimeType == "application/vnd.google-apps.folder"
				childNode := newNode(child.Name, isDir, node)
				childNode.driveID = child.Id
				node.children[child.Name] = childNode
				queue = append(queue, childNode)
			}
		}
	}

	tree := &Tree{
		name: "",
		root: rootNode,
	}
	return tree, nil
}

func (tree *Tree) Root() *Node {
	return tree.root
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
	topPath := tree.RootPath()
	if !strings.HasPrefix(path, topPath) {
		return false
	}
	topPathParts := SplitPathPlatform(topPath)
	pathParts := SplitPathPlatform(path)

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
		childNode := newNode(parts[0], thisPartIsDir, node)
		node.children[parts[0]] = childNode
		addPath(childNode, parts[1:])
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
	topPath := tree.RootPath()
	if !strings.HasPrefix(path, topPath) {
		return nil, false
	}
	topPathParts := SplitPathPlatform(topPath)
	pathParts := SplitPathPlatform(path)

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

// DeletePath removes a path from the tree is present and returns true
// Else returns false
func (tree *Tree) DeletePath(path string) bool {
	node, found := tree.findPath(path)
	if !found {
		return false
	}
	parent := node.parentNode
	delete(parent.children, node.name)
	node = nil
	return true
}

// RenamePath renames an old path to a new path
// The newPath should actually rename the thing (file/folder) referred to by oldPath
// The two paths should thus differ only by the last "element" in the path
// If the rename succeeds, then returns true. else false
func (tree *Tree) RenamePath(oldPath, newPath string) bool {
	oldPathParts := SplitPathPlatform(oldPath)
	newPathParts := SplitPathPlatform(newPath)

	if len(oldPathParts) != len(newPathParts) {
		return false
	}
	for i := 0; i < len(newPathParts)-1; i++ {
		if newPathParts[i] != oldPathParts[i] {
			return false
		}
	}

	node, ok := tree.findPath(oldPath)
	if !ok {
		return false
	}
	node.name = newPathParts[len(newPathParts)-1]
	delete(node.parentNode.children, oldPathParts[len(oldPathParts)-1])
	node.parentNode.children[node.name] = node
	return true
}

// RootPath returns path of root of tree
func (tree *Tree) RootPath() string {
	return filepath.Join(tree.name, tree.root.name)
}

// IsDir returns whether the given path is a directory
// Returns error if path is not found in the tree
func (tree *Tree) IsDir(path string) (bool, error) {
	node, ok := tree.findPath(path)
	if !ok {
		return false, errors.New("Path not found: " + path)
	}
	return node.isDir, nil
}

// ContainsPath returns if the tree contains the given path
func (tree *Tree) ContainsPath(path string) bool {
	_, ok := tree.findPath(path)
	return ok
}

// AttachID attaches the Google Drive id to a file
func (tree *Tree) AttachID(path, id string) bool {
	node, ok := tree.findPath(path)
	if !ok {
		return false
	}
	node.driveID = id
	return true
}

// EqualsIgnore compares two AFS trees, and checks for structural equality
// It provides an option for ignoring the inequality of the root names
func (tree *Tree) EqualsIgnore(other *Tree, ignoreRootName bool) bool {
	return tree.root.EqualsIgnore(other.root, ignoreRootName, false)
}

// Equals compares two AFS trees, and checks for structural equality
func (tree *Tree) Equals(other *Tree) bool {
	return tree.EqualsIgnore(other, false)
}

// EqualsIgnore checks for if the node has same name and the child nodes are the same
// It provides an option to ignore the inequality of names at this level
// It provides an option to specify if the ignoring is propagated down the levels
func (node *Node) EqualsIgnore(other *Node, ignoreName, ignorePropagate bool) bool {
	if !ignoreName && node.name != other.name {
		return false
	}
	for name := range node.children {
		thisChild := node.children[name]
		if otherChild, ok := other.children[name]; !ok {
			return false
		} else {
			ignore := ignoreName && ignorePropagate
			if childEqual := thisChild.EqualsIgnore(otherChild, ignore, ignorePropagate); !childEqual {
				return false
			}
		}
	}
	return len(node.children) == len(other.children)
}
