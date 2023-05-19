//
// Copyright (C) 2023 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/tslab
//

package bst

import (
	"fmt"
	"strings"

	"github.com/fogfish/tslab"
)

type NodeID = tslab.Pointer[Node]

// Node ...
// GC friendly does not contain pointers
type Node struct {
	Value  int
	LH, RH NodeID
	Pad    int
}

// global heap to allocate node objects
var Heap *tslab.Heap[Node] = tslab.New[Node](2)

func (n *Node) Left() *Node  { return n.LH.ValueOf }
func (n *Node) Right() *Node { return n.RH.ValueOf }

func (n *Node) SwapOut(h interface {
	Get(tslab.Pointer[Node]) tslab.Pointer[Node]
}) {
	// fmt.Printf("==> %v %v\n", n.LH, n.RH)

	n.LH = h.Get(n.LH)
	n.RH = h.Get(n.RH)

	// fmt.Printf("==> %v %v\n", n.LH, n.RH)
}

// Recursive tree constructor
func New(value, depth int) NodeID {
	// buf := heap.Alloc()
	// fmt.Printf("%p\n", buf)
	// data := NodeID(buf)
	data := Heap.Alloc()
	node := data.ValueOf
	node.Value = value
	node.Pad = 0xf0f0f0f0

	if depth > 0 {
		node.LH = New(2*value-1, depth-1)
		node.RH = New(2*value+1, depth-1)
	}

	return data
}

// tree traversal algorithm
func Fold(n *Node) int {
	if n == nil {
		return 0
	}

	lh := Fold(n.Left())
	rh := Fold(n.Right())

	return n.Value + lh + rh
}

// tree traversal algorithm
func Count(n *Node) int {
	if n == nil {
		return 0
	}

	lh := Count(n.Left())
	rh := Count(n.Right())

	return 1 + lh + rh
}

func Dump() {
	Heap.Dump()
}

// tree printing
func Print(sb *strings.Builder, pad int, n *Node) error {
	if n == nil {
		return nil
	}

	if err := Print(sb, pad+4, n.Left()); err != nil {
		return err
	}

	_, err := sb.WriteString(fmt.Sprintf("%s%3d\n",
		strings.Repeat(" ", pad), n.Value))

	if err != nil {
		return err
	}

	if err := Print(sb, pad+4, n.Right()); err != nil {
		return err
	}

	return nil
}
