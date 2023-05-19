package main

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/fogfish/tslab"
	"github.com/fogfish/tslab/example/bst"
)

func main() {
	heap := tslab.New[bst.Node](8)
	tree := bst.New(heap, 0, 15)

	fmt.Printf("==> build tree %d\n", bst.Count(tree.ValueOf))

	fd, err := os.Create("/tmp/tslab-full-heap.gob.gz")
	if err != nil {
		panic(err)
	}

	gz := gzip.NewWriter(fd)

	enc := gob.NewEncoder(gz)

	if err := enc.Encode(tree); err != nil {
		panic(err)
	}

	if err := enc.Encode(heap); err != nil {
		panic(err)
	}

	if err := gz.Close(); err != nil {
		panic(err)
	}
}
