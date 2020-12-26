package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// WatchLoop watches for file change events in a loop
// recursive causes it to add new directories being created
func WatchLoop(watcher *fsnotify.Watcher, recursive bool, events chan Event) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Printf("exiting from watch loop")
				return
			}
			fmt.Println(event)

			path := event.Name
			var category EventCategory
			pushEvent := true

			switch event.Op {
			case fsnotify.Create:
				if isDir(path) {
					if recursive {
						AddDirRecursive(path, watcher)
					}
					category = DirectoryCreated
				} else {
					category = FileCreated
				}
			case fsnotify.Remove:
				if isDir(path) {
					category = DirectoryDeleted
				} else {
					category = FileDeleted
				}
			case fsnotify.Write:
				category = FileWritten
			case fsnotify.Rename:
				if isDir(path) {
					category = DirectoryRenamed
				} else {
					category = FileRenamed
				}
			default:
				pushEvent = false
			}

			if pushEvent {
				events <- Event{
					Path:     path,
					Category: category,
				}
			}
		case event, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println(event)
		}
	}
}

func isDir(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

// AddDirRecursive adds directories recursively to watcher
func AddDirRecursive(dir string, watcher *fsnotify.Watcher) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
}
