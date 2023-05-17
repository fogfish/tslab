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

type NodeID = tslab.Ptr[Node]

// Node ...
// GC friendly does not contain pointers
type Node struct {
	left, right NodeID
	value       int
}

// global heap to allocate node objects
var heap *tslab.Heap[Node] = tslab.New[Node](64 * 1024)

func (n *Node) Left() *Node  { return heap.Get(n.left) }
func (n *Node) Right() *Node { return heap.Get(n.right) }

// Recursive tree constructor
func New(value, depth int) (NodeID, *Node) {
	nptr, node := heap.Alloc()
	node.value = value

	if depth > 0 {
		node.left, _ = New(2*value-1, depth-1)
		node.right, _ = New(2*value+1, depth-1)
	}

	return nptr, node
}

// tree traversal algorithm
func Fold(n *Node) int {
	if n == nil {
		return 0
	}

	lh := Fold(n.Left())
	rh := Fold(n.Right())

	return n.value + lh + rh
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

// tree printing
func Print(sb *strings.Builder, pad int, n *Node) error {
	if n == nil {
		return nil
	}

	if err := Print(sb, pad+4, n.Left()); err != nil {
		return err
	}

	_, err := sb.WriteString(fmt.Sprintf("%s%3d\n",
		strings.Repeat(" ", pad), n.value))

	if err != nil {
		return err
	}

	if err := Print(sb, pad+4, n.Right()); err != nil {
		return err
	}

	return nil
}
