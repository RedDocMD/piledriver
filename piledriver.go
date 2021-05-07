package main

import (
	"log"

	"github.com/RedDocMD/piledriver/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
