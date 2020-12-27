package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/RedDocMD/Piledriver/config"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/api/drive/v3"
)

// FileType denotes the file type of a path
type FileType int

// Various sorts of filetypes
const (
	RegularFile FileType = iota
	Directory
)

// State holds global state info for the program
type State struct {
	Config       config.Config
	LogFilePath  string
	FileEvents   chan Event
	watcher      *fsnotify.Watcher
	pathID       map[string]string
	pathType     map[string]FileType
	dirRecursive map[string]bool
	service      *drive.Service
}

// NewState returns a new blank state
func NewState() *State {
	return &State{
		pathID:       make(map[string]string),
		pathType:     make(map[string]FileType),
		dirRecursive: make(map[string]bool),
		FileEvents:   make(chan Event, 512),
	}
}

// InitService initializes the service field
func (state *State) InitService() {
	if state.service == nil {
		state.service = RetrieveDriveService()
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
func (state *State) AddID(path, id string) bool {
	if _, ok := state.pathID[path]; ok {
		return false
	}
	state.pathID[path] = id
	return true
}

func (state *State) scanDir(dir string, recursive bool) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print("Failed to scan - ", err)
			return nil
		}
		var fileType FileType
		if !info.IsDir() {
			fileType = RegularFile
		} else {
			fileType = Directory
			if recursive {
				state.dirRecursive[path] = true
			} else {
				state.dirRecursive[path] = false
			}
		}
		state.pathType[path] = fileType
		if path != dir && info.IsDir() && !recursive {
			return filepath.SkipDir
		}
		return nil
	})
}

// AddDir adds a directory to the watcher and scans paths
func (state *State) AddDir(dir string, recursive bool) {
	state.scanDir(dir, recursive)
	if !recursive {
		state.watcher.Add(dir)
	} else {
		addDirRecursive(dir, state.watcher)
	}
}

// isDir checks a path if it is a directory
// First, it checks its own cache for info
// Then, it tries to check it from the OS
// If it still can't tell, it will return an error
func (state *State) isDir(path string) (bool, error) {
	if val, ok := state.pathType[path]; ok {
		return val == Directory, nil
	} else if info, err := os.Stat(path); err == nil {
		return info.IsDir(), nil
	}
	return false, errors.New(fmt.Sprint(path, ": can't determine if path is a directory"))
}

func (state *State) isDirRecursive(path string) bool {
	return state.dirRecursive[path]
}

func (state *State) addFile(path string) {
	state.pathType[path] = RegularFile
}

func (state *State) delDir(path string) {
	delete(state.dirRecursive, path)
	delete(state.pathType, path)
}

func (state *State) delFile(path string) {
	delete(state.pathType, path)
}
