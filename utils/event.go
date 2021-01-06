package utils

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
