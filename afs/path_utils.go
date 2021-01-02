package afs

import (
	"regexp"
	"strings"
)

// Splits paths into its directories
func splitPath(path string, sep string) []string {
	var partitions []string
	driveRegex := regexp.MustCompile(`[A-Z]:\\`)
	if drive := driveRegex.FindString(path); drive != "" {
		partitions = append(partitions, drive[:2])
		path = path[4:]
	}
	parts := strings.Split(path, sep)
	var nonEmptyParts []string
	for _, part := range parts {
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}
	partitions = append(partitions, nonEmptyParts...)
	return partitions
}

// Reverse of splitPath
func joinPath(parts []string, sep string, isAbs bool) string {
	if len(parts) == 0 {
		return ""
	}
	partsClone := make([]string, len(parts))
	copy(partsClone, parts)
	driveRegex := regexp.MustCompile(`[A-Z]:`)
	if drive := driveRegex.FindString(partsClone[0]); drive != "" {
		partsClone[0] = partsClone[0] + "\\"
	} else if isAbs {
		partsClone[0] = "/" + partsClone[0]
	}
	return strings.Join(partsClone, sep)
}
