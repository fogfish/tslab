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

// NodeID is Pointer type to Node object
type NodeID = tslab.Pointer[Node]

// Node is recursive data structure defining the binary tree
type Node struct {
	Value  int
	LH, RH NodeID
}

func (n *Node) Reloc(heap tslab.Relocator[Node]) {
	n.LH = heap.Reloc(n.LH)
	n.RH = heap.Reloc(n.RH)
}

// Recursive tree constructor
func New(heap tslab.Allocator[Node], value, depth int) NodeID {
	addr := heap.Alloc()
	node := addr.ValueOf
	node.Value = value

	if depth > 0 {
		node.LH = New(heap, 2*value-1, depth-1)
		node.RH = New(heap, 2*value+1, depth-1)
	}

	return addr
}

// Free allocated memory
func Free(heap tslab.Allocator[Node], n NodeID) {
	if n.IsNil() {
		return
	}

	Free(heap, n.ValueOf.LH)
	Free(heap, n.ValueOf.RH)

	heap.Free(n)
}

// tree traversal algorithm - folding tree to single value
func Fold(n *Node) int {
	if n == nil {
		return 0
	}

	lh := Fold(n.LH.ValueOf)
	rh := Fold(n.RH.ValueOf)

	return n.Value + lh + rh
}

// tree traversal algorithm - counting nodes
func Count(n *Node) int {
	if n == nil {
		return 0
	}

	lh := Count(n.LH.ValueOf)
	rh := Count(n.RH.ValueOf)

	return 1 + lh + rh
}

// tree traversal algorithm - converting tree to string
func Print(n *Node) string {
	sb := strings.Builder{}

	if err := print(&sb, "", "    ", n); err != nil {
		return ""
	}

	return sb.String()
}

func print(sb *strings.Builder, pad, indent string, n *Node) error {
	if n == nil {
		return nil
	}

	if err := print(sb, pad+indent, indent, n.LH.ValueOf); err != nil {
		return err
	}

	_, err := sb.WriteString(fmt.Sprintf("%s%3d\n", pad, n.Value))

	if err != nil {
		return err
	}

	if err := print(sb, pad+indent, indent, n.RH.ValueOf); err != nil {
		return err
	}

	return nil
}
