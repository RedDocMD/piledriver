package main

import (
	"github.com/RedDocMD/Piledriver/utils"
)

func main() {
	state := utils.NewState()
	state.InitWatcher()
	state.AddDir("/home/deep/work/dump", true)
	utils.WatchLoop(state)
}
