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
	fd, err := os.Open("/tmp/tslab-full-heap.gob.gz")
	if err != nil {
		panic(err)
	}

	gz, err := gzip.NewReader(fd)
	if err != nil {
		panic(err)
	}

	dec := gob.NewDecoder(gz)

	heap := tslab.New[bst.Node](8)
	var tree bst.NodeID

	if err := dec.Decode(&tree); err != nil {
		panic(err)
	}

	if err := dec.Decode(heap); err != nil {
		panic(err)
	}

	tree = heap.Reloc(tree)

	fmt.Printf("==> re-build tree %d\n", bst.Count(tree.ValueOf))
}
