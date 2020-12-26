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
func WatchLoop(watcher *fsnotify.Watcher, recursive bool) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			fmt.Println(event)
			if event.Op == fsnotify.Create {
				if isDir(event.Name) {
					AddDirRecursive(event.Name, watcher)
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
