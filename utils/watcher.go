package utils

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchLoop watches for file change events in a loop
// recursive causes it to add new directories being created
func WatchLoop(state *State) {
	watcher := state.watcher
	events := state.FileEvents

	renamePending := false
	pathToBeRenamed := ""

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("exiting from watch loop")
				return
			}

			path := event.Name
			var category EventCategory
			pushEvent := true

			idMap := make(map[IDKey]string)
			timestamp := time.Now()

			var isDir bool
			var err error
			if !renamePending {
				isDir, err = state.isDir(path)
				if err != nil {
					stat, err := os.Stat(path)
					if os.IsNotExist(err) {
						log.Println("Failed to open:", path)
					} else {
						isDir = stat.IsDir()
					}
				}
			} else {
				isDir, err = state.isDir(pathToBeRenamed)
				if err != nil {
					pushEvent = false
				}
			}

			switch event.Op {
			case fsnotify.Create:
				if renamePending {
					if ok := state.renamePath(pathToBeRenamed, path); !ok {
						log.Printf("Cannot rename %s to %s", pathToBeRenamed, path)
						pushEvent = false
					}
					if isDir {
						category = DirectoryRenamed
					} else {
						category = FileRenamed
					}
					renamePending = false
				} else {
					if isDir {
						err := state.AddDir(path)
						if err != nil {
							log.Println("Failed to add", path, "to tree")
						}
						category = DirectoryCreated
					} else {
						ok := state.addFile(path)
						if !ok {
							log.Println("Failed to add", path, "to tree")
						}
						category = FileCreated
					}
				}
			case fsnotify.Remove:
				if isDir {
					category = DirectoryDeleted
				} else {
					category = FileDeleted
				}
				id, ok := state.retrieveID(path)
				if !ok {
					log.Printf("Cannot retrieve ID of %s\n", path)
					pushEvent = false
				} else {
					idMap[CurrID] = id
					if ok := state.delPath(path); !ok {
						log.Println("Cannot delete: ", path)
						pushEvent = false
					}
				}
			case fsnotify.Write:
				category = FileWritten
			case fsnotify.Rename:
				pushEvent = false
				if state.pathExists(path) {
					renamePending = true
					pathToBeRenamed = path
				}
			default:
				pushEvent = false
			}

			if pushEvent {
				events <- Event{
					Path:      path,
					OldPath:   pathToBeRenamed,
					Category:  category,
					IDMap:     idMap,
					Timestamp: timestamp,
				}
			}
		case event, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error: ", event)
		}
	}
}

func addDirRecursive(dir string, watcher *fsnotify.Watcher) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print("Failed to add - ", err)
			return err
		}
		if info.IsDir() {
			err := watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
