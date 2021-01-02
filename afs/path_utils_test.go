package afs

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestEmptyPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := ""
	unixPartitions := splitPath(path, "/")
	windowsPartitions := splitPath(path, "\\")
	assert.Equal(0, len(unixPartitions))
	assert.Equal(0, len(windowsPartitions))
}

func TestEmptyPathJoin(t *testing.T) {
	assert := assert.New(t)
	var parts []string
	unixPath := joinPath(parts, "/", true)
	windowsPath := joinPath(parts, "\\", true)
	assert.Equal("", unixPath)
	assert.Equal("", windowsPath)
}

// Unix path tests

func TestUnixAbsPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := "/home/joe/learns/to/walk"
	parts := splitPath(path, "/")
	partsExp := []string{"home", "joe", "learns", "to", "walk"}
	assert.Equal(len(partsExp), len(parts))
	for i, part := range parts {
		assert.Equal(partsExp[i], part)
	}
}

func TestUnixAbsPathJoin(t *testing.T) {
	assert := assert.New(t)
	parts := []string{"home", "joe", "learns", "to", "walk"}
	path := joinPath(parts, "/", true)
	pathExp := "/home/joe/learns/to/walk"
	assert.Equal(pathExp, path)
}

func TestUnixAbsDirPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := "/home/joe/learns/to/walk/"
	parts := splitPath(path, "/")
	partsExp := []string{"home", "joe", "learns", "to", "walk"}
	assert.Equal(len(partsExp), len(parts))
	for i, part := range parts {
		assert.Equal(partsExp[i], part)
	}
}

func TestUnixRelPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := "joe/learns/to/walk"
	parts := splitPath(path, "/")
	partsExp := []string{"joe", "learns", "to", "walk"}
	assert.Equal(len(partsExp), len(parts))
	for i, part := range parts {
		assert.Equal(partsExp[i], part)
	}
}

func TestUnixRelPathJoin(t *testing.T) {
	assert := assert.New(t)
	parts := []string{"joe", "learns", "to", "walk"}
	path := joinPath(parts, "/", false)
	pathExp := "joe/learns/to/walk"
	assert.Equal(pathExp, path)
}

func TestUnixRelDirPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := "joe/learns/to/walk/"
	parts := splitPath(path, "/")
	partsExp := []string{"joe", "learns", "to", "walk"}
	assert.Equal(len(partsExp), len(parts))
	for i, part := range parts {
		assert.Equal(partsExp[i], part)
	}
}

// Windows path tests

func TestWindowsAbsPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := `C:\\home\joe\learns\to\walk`
	parts := splitPath(path, "\\")
	partsExp := []string{"C:", "home", "joe", "learns", "to", "walk"}
	assert.Equal(len(partsExp), len(parts))
	for i, part := range parts {
		assert.Equal(partsExp[i], part)
	}
}

func TestWindowsAbsPathJoin(t *testing.T) {
	assert := assert.New(t)
	parts := []string{"C:", "home", "joe", "learns", "to", "walk"}
	path := joinPath(parts, "\\", true)
	pathExp := `C:\\home\joe\learns\to\walk`
	assert.Equal(pathExp, path)
}

func TestWindowsAbsDirPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := `C:\\home\joe\learns\to\walk\`
	parts := splitPath(path, "\\")
	partsExp := []string{"C:", "home", "joe", "learns", "to", "walk"}
	assert.Equal(len(partsExp), len(parts))
	for i, part := range parts {
		assert.Equal(partsExp[i], part)
	}
}

func TestWindowsRelPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := `joe\learns\to\walk`
	parts := splitPath(path, "\\")
	partsExp := []string{"joe", "learns", "to", "walk"}
	assert.Equal(len(partsExp), len(parts))
	for i, part := range parts {
		assert.Equal(partsExp[i], part)
	}
}

func TestWindowsRelPathJoin(t *testing.T) {
	assert := assert.New(t)
	parts := []string{"home", "joe", "learns", "to", "walk"}
	path := joinPath(parts, "\\", false)
	pathExp := `home\joe\learns\to\walk`
	assert.Equal(pathExp, path)
}

func TestWindowsRelDirPathSplit(t *testing.T) {
	assert := assert.New(t)
	path := `joe\learns\to\walk\`
	parts := splitPath(path, "\\")
	partsExp := []string{"joe", "learns", "to", "walk"}
	assert.Equal(len(partsExp), len(parts))
	for i, part := range parts {
		assert.Equal(partsExp[i], part)
	}
}
