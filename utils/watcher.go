package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
					parts := strings.Split(path, string(filepath.Separator))
					parentDir := "/" + filepath.Join(parts[:len(parts)-1]...)
					if state.isDirRecursive(parentDir) {
						state.AddDir(path, true)
					}
					category = DirectoryCreated
				} else {
					state.addFile(path)
					category = FileCreated
				}
			case fsnotify.Remove:
				if isDir {
					category = DirectoryDeleted
					state.delDir(path)
				} else {
					category = FileDeleted
					state.delFile(path)
				}
			case fsnotify.Write:
				category = FileWritten
				// TODO: Put state updating
			case fsnotify.Rename:
				// TODO: Put state updating
				if isDir {
					category = DirectoryRenamed
					// Remove all old paths
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
