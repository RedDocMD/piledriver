package main

import (
	"github.com/RedDocMD/piledriver/utils"
)

func main() {
	state := utils.NewState()
	state.InitService()
	state.InitWatcher()
	state.AddDir("/home/dknite/work/dump", true)
	go utils.ExecuteEvents(state.FileEvents)
	utils.WatchLoop(state)
}
