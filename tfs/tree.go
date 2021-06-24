package tfs

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

const ChecksumSize uint = 20

type Node struct {
	name     string
	parent   *Node
	children []*Node
	isDir    bool
	checksum [ChecksumSize]byte
	handle   *os.File
	stat     *unix.Stat_t
}

// NewNode creates an node representing a Unix file.
// It is un-initialized. InitNode must be called.
func NewNode(name string, isDir bool) *Node {
	return &Node{
		name:  name,
		isDir: isDir,
	}
}

// InitNode initializes a node by filling in the
// checksum, stat and handle.
// This operation can fail, either due to lack of
// sufficient permissions, or the file is missing.
func InitNode(node *Node, path string) error {
	node.stat = new(unix.Stat_t)
	err := unix.Stat(path, node.stat)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}
	node.handle, err = os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}
	if !node.isDir {
		var buf bytes.Buffer
		_, err = io.Copy(&buf, node.handle)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}
		sum := sha1.Sum(buf.Bytes())
		copy(node.checksum[:], sum[:])
	}
	return nil
}

type Tree struct {
	rootPath string
	root     *Node
}

// Creates the Tree, recursively adding children
// If it fails at the root node, it returns an error
// It if fails at any other file/folder, it is simply IGNORED.
func TreeFromPath(path string) (*Tree, error) {
	slashIdx := strings.LastIndex(path[0:len(path)-1], "/")
	rootPath := path[0 : slashIdx+1]
	rootName := path[slashIdx+1:]
	rootNode := NewNode(rootName, true)
	err := InitNode(rootNode, path)
	if err != nil {
		return nil, fmt.Errorf("failed to init root node %s: %w", path, err)
	}

	var addChildren func(path string, node *Node)
	addChildren = func(path string, node *Node) {
		if !node.isDir {
			return
		}
		files, err := node.handle.Readdir(-1)
		if err != nil {
			return
		}
		for _, file := range files {
			newNode := NewNode(file.Name(), file.IsDir())
			newPath := filepath.Join(path, file.Name())
			err = InitNode(newNode, newPath)
			if err != nil {
				continue
			}
			node.children = append(node.children, newNode)
			newNode.parent = node
			addChildren(newPath, newNode)
		}
	}

	addChildren(path, rootNode)

	return &Tree{rootPath: rootPath, root: rootNode}, nil
}
