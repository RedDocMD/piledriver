package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/RedDocMD/piledriver/afs"
)

// EventCategory denotes the type of event that has been detected
type EventCategory uint

// Various types of event categories
const (
	FileCreated EventCategory = iota
	DirectoryCreated
	FileDeleted
	DirectoryDeleted
	FileRenamed
	DirectoryRenamed
	FileWritten
)

// IDKey denotes the key for the ID type
type IDKey int

// Various types of id's
const (
	CurrID IDKey = iota
	ParentID
)

// Event is the internal representation of file watcher events
type Event struct {
	OldPath  string
	Path     string
	Category EventCategory
	IDMap    map[IDKey]string
}

func (ev Event) String() string {
	var catString string
	switch ev.Category {
	case FileCreated:
		catString = "FILE CREATED"
	case DirectoryCreated:
		catString = "DIRECTORY CREATED"
	case FileDeleted:
		catString = "FILE DELETED"
	case DirectoryDeleted:
		catString = "DIRECTORY DELETED"
	case FileRenamed:
		catString = "FILE RENAMED"
	case DirectoryRenamed:
		catString = "DIRECTORY RENAMED"
	case FileWritten:
		catString = "FILE WRITTEN"
	default:
		catString = "Unknown event type"
	}

	if ev.Category == DirectoryRenamed || ev.Category == FileRenamed {
		return fmt.Sprintf("%s   %s => %s", catString, ev.OldPath, ev.Path)
	}
	return fmt.Sprintf("%s   %s", catString, ev.Path)
}

// ExecuteEvents takes a channel Events and executes them
func ExecuteEvents(state *State) {
	var getID, getParentID func(path string) (string, bool)
	var pathName func(path string) string

	getID = func(path string) (string, bool) {
		return state.retrieveID(path)
	}
	getParentID = func(path string) (string, bool) {
		pathParts := afs.SplitPathPlatform(path)
		parentPath := afs.JoinPathPlatform(pathParts[0:len(pathParts)-1], true)
		return state.retrieveID(parentPath)
	}
	pathName = func(path string) string {
		parts := afs.SplitPathPlatform(path)
		return parts[len(parts)-1]
	}

	const sleepTime = 1 * time.Second

	for ev := range state.FileEvents {
		log.Println(ev)
		switch ev.Category {
		case FileCreated:
			path := ev.Path
			parentID, ok := getParentID(path)
			if !ok {
				log.Printf("Node for parent of %s not found\n", path)
				continue
			}
			var fileID string
			var err error
			for {
				fileID, err = CreateFile(state.service, path, parentID)
				if err != nil && err.Error() != "Failed in file IO" {
					log.Println(err)
					time.Sleep(sleepTime)
				} else {
					if err != nil {
						log.Printf("Failed to read from %s\n", path)
					}
					break
				}
			}
			ok = state.attachID(path, fileID)
			if !ok {
				log.Printf("Failed to attach id of %s\n", path)
			}
		case DirectoryCreated:
			path := ev.Path
			parentID, ok := getParentID(path)
			if !ok {
				log.Printf("Node for parent of %s not found\n", path)
				continue
			}
			var fileID string
			var err error
			for {
				fileID, err = CreateFolder(state.service, path, parentID)
				if err != nil {
					log.Println(err)
					time.Sleep(sleepTime)
				} else {
					break
				}
			}
			ok = state.attachID(path, fileID)
			if !ok {
				log.Printf("Failed to attach id of %s\n", path)
			}
		case FileDeleted:
			fallthrough
		case DirectoryDeleted:
			id := ev.IDMap[CurrID]
			for {
				err := DeleteFileOrFolder(state.service, id)
				if err != nil {
					log.Println(err)
					time.Sleep(sleepTime)
				} else {
					break
				}
			}
		case FileRenamed:
			fallthrough
		case DirectoryRenamed:
			oldPath := ev.OldPath
			newPath := ev.Path

			id, ok := getID(newPath)
			if !ok {
				log.Printf("Failed to retrieve ID of %s\n", newPath)
				continue
			}
			oldParentID, ok := getParentID(oldPath)
			if !ok {
				log.Printf("Failed to retrieve ID of parent of %s\n", oldPath)
				continue
			}
			newParentID, ok := getParentID(newPath)
			if !ok {
				log.Printf("Failed to retrieve ID of parent of %s\n", newPath)
				continue
			}

			info := RenameInfo{
				ID:          id,
				NewParentID: newParentID,
				OldParentID: oldParentID,
				NewName:     pathName(newPath),
			}

			for {
				_, err := RenameFileOrFolder(state.service, info)
				if err != nil {
					log.Println(err)
					time.Sleep(sleepTime)
				} else {
					break
				}
			}
		case FileWritten:
			path := ev.Path
			id, ok := getID(path)
			if !ok {
				log.Printf("Failed to retrieve ID of %s\n", path)
				continue
			}
			for {
				_, err := UpdateFile(state.service, path, id)
				if err != nil && err.Error() != "Failed in file IO" {
					log.Println(err)
					time.Sleep(sleepTime)
				} else {
					if err != nil {
						log.Printf("Failed to read from %s\n", path)
					}
					break
				}

			}
		}
	}
}
