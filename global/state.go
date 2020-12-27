package global

import (
	"github.com/RedDocMD/Piledriver/config"
	"github.com/RedDocMD/Piledriver/utils"
	"google.golang.org/api/drive/v3"
)

// State holds global state info for the program
type State struct {
	Config      config.Config
	LogFilePath string
	FileEvents  chan utils.Event
	pathID      map[string]string
	service     *drive.Service
}

// NewState returns a new blank state
func NewState() *State {
	return &State{
		pathID:     make(map[string]string),
		FileEvents: make(chan utils.Event, 512),
	}
}

// InitService initializes the service field
func (state *State) InitService() {
	if state.service == nil {
		state.service = utils.RetrieveDriveService()
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
