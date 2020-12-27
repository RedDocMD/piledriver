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
func WatchLoop(state *State) {
	watcher := state.watcher
	events := state.FileEvents

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
			isDir, err := state.isDir(path)
			if err != nil {
				log.Fatal(err)
			}

			switch event.Op {
			case fsnotify.Create:
				if isDir {
					state.AddDir(path, true) // TODO: Correct this true
					category = DirectoryCreated
				} else {
					category = FileCreated
				}
			case fsnotify.Remove:
				if isDir {
					category = DirectoryDeleted
				} else {
					category = FileDeleted
				}
			case fsnotify.Write:
				category = FileWritten
			case fsnotify.Rename:
				if isDir {
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

func addDirRecursive(dir string, watcher *fsnotify.Watcher) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print("Failed to add - ", err)
			return nil
		}
		if info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
}
