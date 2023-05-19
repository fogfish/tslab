//
// Copyright (C) 2023 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/tslab
//

package main

import (
	"fmt"

	"github.com/fogfish/tslab"
	"github.com/fogfish/tslab/example/bst"
)

func main() {
	heap := tslab.New[bst.Node](8)
	tree := bst.New(heap, 0, 4)

	fmt.Println(bst.Print(tree.ValueOf))
	fmt.Printf("==> total nodes 2ʰ⁺¹-1: %d\n", bst.Count(tree.ValueOf))

	fmt.Printf("==> stats %+3v\n", heap.Stats())

	bst.Free(heap, tree)
	fmt.Printf("==> stats %+3v\n", heap.Stats())
}
