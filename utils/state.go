package utils

import (
	"log"

	"github.com/RedDocMD/Piledriver/config"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/api/drive/v3"
)

// State holds global state info for the program
type State struct {
	Config      config.Config
	LogFilePath string
	FileEvents  chan Event
	watcher     *fsnotify.Watcher
	pathID      map[string]string
	service     *drive.Service
}

// NewState returns a new blank state
func NewState() *State {
	return &State{
		pathID:     make(map[string]string),
		FileEvents: make(chan Event, 512),
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

// Watcher returns the watcher field
func (state *State) Watcher() *fsnotify.Watcher {
	return state.watcher
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
