package afs

import (
	"strings"
)

// Splits paths into its directories
func splitPath(path string, sep string) []string {
	partions := strings.Split(path, sep)
	var newPartitions []string
	for _, val := range partions {
		if val != "" {
			newPartitions = append(newPartitions, val)
		}
	}
	return newPartitions
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
