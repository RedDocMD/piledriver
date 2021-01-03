package afs

import (
	"fmt"
	"log"
	"path/filepath"
)

func ExampleNewTree() {
	path, err := filepath.Abs("test_data/rec_dir")
	if err != nil {
		log.Fatal(err)
	}
	tree := NewTree(path, true)
	fmt.Println(tree)
}
