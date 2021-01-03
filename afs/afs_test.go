package afs

import "fmt"

func ExampleNewTree() {
	tree := NewTree("test_data/rec_dir", true, false)
	fmt.Println(tree)
}
