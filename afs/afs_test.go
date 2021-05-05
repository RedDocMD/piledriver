package afs

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"
)

func extendNode(node *Node, currPath string) {
	if node.isDir {
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
			newPath := filepath.Join(currPath, name)
			if stat, err := os.Stat(newPath); !os.IsNotExist(err) {
				newIsDir := stat.IsDir()
				newNode := newNode(name, newIsDir, node)
				node.children[name] = newNode
				extendNode(newNode, newPath)
			}
		}
	}
}

func constructTree() *Tree {
	path, err := filepath.Abs(filepath.FromSlash("test_data/rec_dir"))
	if err != nil {
		log.Fatal(err)
	}
	tree := NewTree(path)
	extendNode(tree.root, "test_data/rec_dir")

	return tree
}

func ExampleNewTree() {
	tree := constructTree()
	log.Println(tree)
}

func TestFindPath(t *testing.T) {
	assert := assert.New(t)
	path, _ := filepath.Abs(filepath.FromSlash("test_data/rec_dir"))
	tree := constructTree()

	_, found := tree.findPath(path)
	assert.True(found)

	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir1/dir3/file6")))
	assert.True(found)

	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir2/file7")))
	assert.True(found)

	_, found = tree.findPath(filepath.Join(path, "file3"))
	assert.True(found)

	_, found = tree.findPath(filepath.Join(path, "file10"))
	assert.False(found)

	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir11/file10")))
	assert.False(found)

	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir1/dir3/file22")))
	assert.False(found)
}

func TestAddPath(t *testing.T) {
	assert := assert.New(t)
	tempDir := os.TempDir()
	basePath := filepath.Join(tempDir, "test")

	tree := NewTree(basePath)

	newPath := filepath.Join(basePath, filepath.FromSlash("dir1/dir2"))
	tree.AddPath(newPath, true)
	node, ok := tree.findPath(newPath)
	assert.True(ok)
	assert.True(node.isDir)
	assert.Equal(node.name, "dir2")

	newPath = filepath.Join(basePath, filepath.FromSlash("dir1/file"))
	tree.AddPath(newPath, false)
	node, ok = tree.findPath(newPath)
	assert.True(ok)
	assert.False(node.isDir)
	assert.Equal(node.name, "file")
}

func TestDeletePath(t *testing.T) {
	path, _ := filepath.Abs(filepath.FromSlash("test_data/rec_dir"))
	assert := assert.New(t)
	tree := constructTree()

	_, found := tree.findPath(filepath.Join(path, filepath.FromSlash("dir1/dir3/file6")))
	assert.True(found)
	tree.DeletePath(filepath.Join(path, filepath.FromSlash("dir1/dir3/file6")))
	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir1/dir3/file6")))
	assert.False(found)

	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir2/file7")))
	assert.True(found)
	tree.DeletePath(filepath.Join(path, "dir2"))
	_, found = tree.findPath(filepath.Join(path, "dir2"))
	assert.False(found)
	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir2/file7")))
	assert.False(found)
}

func TestRenamePath(t *testing.T) {
	path, _ := filepath.Abs(filepath.FromSlash("test_data/rec_dir"))
	assert := assert.New(t)
	tree := constructTree()

	_, found := tree.findPath(filepath.Join(path, filepath.FromSlash("dir1/dir3/file6")))
	assert.True(found)
	tree.RenamePath(filepath.Join(path, filepath.FromSlash("dir1/dir3/file6")),
		filepath.Join(path, filepath.FromSlash("dir1/dir3/newfile")))
	_, found = tree.findPath(filepath.Join(path, "dir1/dir3/newfile"))
	assert.True(found)
	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir1/dir3/file6")))
	assert.False(found)

	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir2/file7")))
	assert.True(found)
	tree.RenamePath(filepath.Join(path, "dir2"), filepath.Join(path, "dirnew"))
	_, found = tree.findPath(filepath.Join(path, "dir2"))
	assert.False(found)
	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dir2/file7")))
	assert.False(found)
	_, found = tree.findPath(filepath.Join(path, "dirnew"))
	assert.True(found)
	_, found = tree.findPath(filepath.Join(path, filepath.FromSlash("dirnew/file7")))
	assert.True(found)
}
