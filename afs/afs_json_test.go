package afs

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
)

func TestEncode(t *testing.T) {
	assert := assert.New(t)
	tree := NewTree("hello", true)
	addPathAndExpect(assert, tree, JoinPathPlatform([]string{"hello", "moron", "file1"}, false), false)
	addPathAndExpect(assert, tree, JoinPathPlatform([]string{"hello", "moron", "file2"}, false), false)
}

func addPathAndExpect(a *assert.Assertions, tree *Tree, path string, isDir bool) {
	done := tree.AddPath(path, isDir)
	a.True(done, fmt.Sprintf("Failed to add %s", path))
}
