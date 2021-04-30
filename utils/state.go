package utils

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/RedDocMD/piledriver/afs"
	"github.com/RedDocMD/piledriver/config"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/api/drive/v3"
)

// State holds global state info for the program
type State struct {
	Config      config.Config
	LogFilePath string
	FileEvents  chan Event
	watcher     *fsnotify.Watcher
	service     *drive.Service
	trees       map[string]*afs.Tree // Map from root path to tree
}

// NewState returns a new blank state
func NewState() *State {
	return &State{
		FileEvents: make(chan Event, 512),
		trees:      make(map[string]*afs.Tree),
	}
}

// InitService initializes the service field
func (state *State) InitService(tokenPath string) {
	if state.service == nil {
		state.service = GetDriveService(tokenPath)
	}
}

// InitWatcher initializes the watcher field
func (state *State) InitWatcher() {
	if state.watcher == nil {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		state.watcher = watcher
	}
}

// Service returns the service field
func (state *State) Service() *drive.Service {
	return state.service
}

// AddID adds in the ID of a path, returning true
// Returns false if it already exists, doesn't overwrite
// func (state *State) AddID(path, id string) bool {
// 	if _, ok := state.pathID[path]; ok {
// 		return false
// 	}
// 	state.pathID[path] = id
// 	return true
// }

func (state *State) scanDir(dir string, recursive bool) {
	// Assume that dir has already been added to state.trees
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print("Failed to scan - ", err)
			return nil
		}
		for name := range state.trees {
			if strings.HasPrefix(path, name) {
				state.trees[name].AddPath(path, info.IsDir())
			}
		}
		if path != dir && info.IsDir() && !recursive {
			return filepath.SkipDir
		}
		return nil
	})
}

// AddDir adds a directory to the watcher and scans paths
func (state *State) AddDir(dir string, recursive bool) {
	added := false
	for name := range state.trees {
		if strings.HasPrefix(dir, name) {
			state.trees[name].AddPath(dir, true)
			added = true
		}
	}
	if !added {
		tree := afs.NewTree(dir, recursive)
		state.trees[tree.Name()] = tree
	}
	state.scanDir(dir, recursive)
	if !recursive {
		state.watcher.Add(dir)
	} else {
		addDirRecursive(dir, state.watcher)
	}
}

// isDir checks a path if it is a directory
func (state *State) isDir(path string) (bool, error) {
	for name := range state.trees {
		if strings.HasPrefix(path, name) {
			stat, err := state.trees[name].IsDir(path)
			if err != nil {
				return false, err
			}
			return stat, nil
		}
	}
	return false, errors.New("Path not found in any tree: " + path)
}

func (state *State) isDirRecursive(path string) (bool, error) {
	for name := range state.trees {
		if strings.HasPrefix(path, name) {
			stat, err := state.trees[name].IsRecursive(path)
			if err != nil {
				return false, err
			}
			return stat, nil
		}
	}
	return false, errors.New("Path not found in any tree: " + path)
}

// Adds a file and returns whether it was actually added
func (state *State) addFile(path string) bool {
	// Assume parent directory has been added before
	for name := range state.trees {
		if strings.HasPrefix(path, name) {
			done := state.trees[name].AddPath(path, false)
			return done
		}
	}
	return false
}

func (state *State) delPath(path string) bool {
	for name := range state.trees {
		if strings.HasPrefix(path, name) {
			done := state.trees[name].DeletePath(path)
			return done
		}
	}
	return false
}

func (state *State) renamePath(oldPath, newPath string) bool {
	for name, tree := range state.trees {
		if strings.HasPrefix(oldPath, name) {
			isDir, _ := tree.IsDir(oldPath)
			isRecursive, _ := tree.IsRecursive(oldPath)
			done := tree.RenamePath(oldPath, newPath)
			if isDir {
				if isRecursive {
					addDirRecursive(newPath, state.watcher)
				} else {
					state.watcher.Add(newPath)
				}
			}
			return done
		}
	}
	return false
}

func (state *State) pathExists(path string) bool {
	for _, tree := range state.trees {
		if tree.ContainsPath(path) {
			return true
		}
	}
	return false
}
