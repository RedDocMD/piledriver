package tfs

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"

	"github.com/alecthomas/assert"
)

func TestNewTree(t *testing.T) {
	assert := assert.New(t)
	rootPath, err := filepath.Abs(filepath.FromSlash("test_data/rec_dir"))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := TreeFromPath(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(len(tree.root.children), 5)
	assert.Equal(tree.root.name, "rec_dir")

	checksums := [9]string{
		"4746e185be32b6a012e763ef04d8493f136b9a1e",
		"f57ac86275e08cddefbba6519e9d54264fa064bf",
		"73f8d4e300cbf9093b857622ccf404be8075d560",
		"127767605c0e978d994939a750e4b4373217dec0",
		"e9203d4eef8bcaa9e621fcb9fcc9de6683361e94",
		"dbb6e2cc3daa833cdc449cea379a4eb2bbbafa99",
		"c86ba2478fd8802248839af52f23ed0ae1d408cf",
		"8042251eaa48ca83dabec1b73edaf2a6f9230451",
		"bd7e45410c30e867041b8c0bf6f2c3bb510669cf",
	}
	re := regexp.MustCompile(`^file\d$`)

	var stack []*Node
	stack = append(stack, tree.root)
	for len(stack) != 0 {
		node := stack[len(stack)-1]
		stack = stack[0 : len(stack)-1]
		if !node.isDir {
			assert.NotEqual(re.FindString(node.name), "")
			idx := int(node.name[len(node.name)-1] - 48)
			assert.Equal(checksums[idx-1], fmt.Sprintf("%x", node.checksum))
		} else {
			stack = append(stack, node.children...)
		}
	}
}

func TestFailedTree(t *testing.T) {
	assert := assert.New(t)
	rootPath, err := filepath.Abs(filepath.FromSlash("test_data/crummy_dir"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = TreeFromPath(rootPath)
	assert.True(errors.Is(err, syscall.ENOENT))
}
