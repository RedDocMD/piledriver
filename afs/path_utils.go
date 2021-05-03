package afs

import (
	"path/filepath"
	"strings"
)

// Splits paths into its directories
func splitPath(path, sep string) []string {
	partions := strings.Split(path, sep)
	var newPartitions []string
	for _, val := range partions {
		if val != "" {
			newPartitions = append(newPartitions, val)
		}
	}
	return newPartitions
}

// SplitPathPlatform splits a path string, letting the stdlib decide the path seperator
func SplitPathPlatform(path string) []string {
	return splitPath(path, string(filepath.Separator))
}

// Reverse of splitPath
func joinPath(parts []string, sep string, isAbs bool) string {
	if len(parts) == 0 {
		return ""
	}
	joined := strings.Join(parts, sep)
	if isAbs && !strings.Contains(parts[0], ":") {
		return "/" + joined
	}
	return joined
}

// JoinPathPlatform joins a path string, letting the stdlib decide the path seperator
func JoinPathPlatform(parts []string, isAbs bool) string {
	return joinPath(parts, string(filepath.Separator), isAbs)
}
