package afs

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"
)

func ExampleNewTree() {
	path, err := filepath.Abs("test_data/rec_dir")
	if err != nil {
		log.Fatal(err)
	}
	tree := NewTree(path, true)
	fmt.Println(tree)
}

func TestFindPath(t *testing.T) {
	path, err := filepath.Abs("test_data/rec_dir")
	assert := assert.New(t)
	if err != nil {
		log.Fatal(err)
	}
	tree := NewTree(path, true)

	_, found := tree.findPath(path)
	assert.True(found)

	_, found = tree.findPath(filepath.Join(path, "dir1/dir3/file6"))
	assert.True(found)

	_, found = tree.findPath(filepath.Join(path, "dir2/file7"))
	assert.True(found)

	_, found = tree.findPath(filepath.Join(path, "file3"))
	assert.True(found)

	_, found = tree.findPath(filepath.Join(path, "file10"))
	assert.False(found)

	_, found = tree.findPath(filepath.Join(path, "dir11/file10"))
	assert.False(found)

	_, found = tree.findPath(filepath.Join(path, "dir1/dir3/file22"))
	assert.False(found)
}

func TestAddPath(t *testing.T) {
	assert := assert.New(t)
	tempDir := os.TempDir()
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		basePath := path.Join(tempDir, "test")
		err = os.Mkdir(basePath, os.ModePerm)
		if err != nil {
			t.Error(err)
		}

		tree := NewTree(basePath, true)

		newPath := path.Join(basePath, "dir1/dir2")
		err = os.MkdirAll(newPath, os.ModePerm)
		if err != nil {
			t.Error(err)
		}
		tree.AddPath(newPath, true)
		node, ok := tree.findPath(newPath)
		assert.True(ok)
		assert.True(node.isDir)
		assert.Equal(node.name, "dir2")

		newPath = path.Join(basePath, "dir1/file")
		file, err := os.Create(newPath)
		defer file.Close()
		if err != nil {
			t.Error(err)
		}
		tree.AddPath(newPath, false)
		node, ok = tree.findPath(newPath)
		assert.True(ok)
		assert.False(node.isDir)
		assert.Equal(node.name, "file")

		os.RemoveAll(basePath)
	} else {
		t.Error("Cannot access tmp directory")
	}
}

func TestDeletePath(t *testing.T) {
	path, err := filepath.Abs("test_data/rec_dir")
	assert := assert.New(t)
	if err != nil {
		log.Fatal(err)
	}
	tree := NewTree(path, true)

	_, found := tree.findPath(filepath.Join(path, "dir1/dir3/file6"))
	assert.True(found)
	tree.DeletePath(filepath.Join(path, "dir1/dir3/file6"))
	_, found = tree.findPath(filepath.Join(path, "dir1/dir3/file6"))
	assert.False(found)

	_, found = tree.findPath(filepath.Join(path, "dir2/file7"))
	assert.True(found)
	tree.DeletePath(filepath.Join(path, "dir2"))
	_, found = tree.findPath(filepath.Join(path, "dir2"))
	assert.False(found)
	_, found = tree.findPath(filepath.Join(path, "dir2/file7"))
	assert.False(found)
}

func TestRenamePath(t *testing.T) {
	path, err := filepath.Abs("test_data/rec_dir")
	assert := assert.New(t)
	if err != nil {
		log.Fatal(err)
	}
	tree := NewTree(path, true)

	_, found := tree.findPath(filepath.Join(path, "dir1/dir3/file6"))
	assert.True(found)
	tree.RenamePath(filepath.Join(path, "dir1/dir3/file6"), filepath.Join(path, "dir1/dir3/newfile"))
	_, found = tree.findPath(filepath.Join(path, "dir1/dir3/newfile"))
	assert.True(found)
	_, found = tree.findPath(filepath.Join(path, "dir1/dir3/file6"))
	assert.False(found)

	_, found = tree.findPath(filepath.Join(path, "dir2/file7"))
	assert.True(found)
	tree.RenamePath(filepath.Join(path, "dir2"), filepath.Join(path, "dirnew"))
	_, found = tree.findPath(filepath.Join(path, "dir2"))
	assert.False(found)
	_, found = tree.findPath(filepath.Join(path, "dir2/file7"))
	assert.False(found)
	_, found = tree.findPath(filepath.Join(path, "dirnew"))
	assert.True(found)
	_, found = tree.findPath(filepath.Join(path, "dirnew/file7"))
	assert.True(found)
}
