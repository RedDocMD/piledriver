package utils

import (
	"fmt"
	"log"
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

// Event is the internal representation of file watcher events
type Event struct {
	OldPath  string
	Path     string
	Category EventCategory
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
		catString = "FILE"
	case DirectoryRenamed:
		catString = "DIRECTORY"
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
func ExecuteEvents(events chan Event) {
	for event := range events {
		log.Println(event)
	}
}
