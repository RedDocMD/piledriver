package main

import (
	"github.com/RedDocMD/Piledriver/utils"
)

func main() {
	state := utils.NewState()
	state.InitWatcher()
	state.AddDir("/tmp", true)
	go utils.WatchLoop(state)
	for {
	}
}
