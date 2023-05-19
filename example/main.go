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
	"os"
	"runtime"
	"strings"

	"github.com/fogfish/tslab"
	"github.com/fogfish/tslab/example/bst"
)

func main() {

	// tree := bst.New(0, 4)

	// sb := strings.Builder{}
	// bst.Print(&sb, 0, tree.ValueOf)
	// fmt.Println(sb.String())

	// b, _ := bst.Heap.DumpX()
	// // fmt.Printf("%v %+x\n", err, b)

	// os.WriteFile("xxx.gob", b, 0777)
	// fmt.Println(tree)

	dat, _ := os.ReadFile("xxx.gob")

	err := bst.Heap.UnDump(dat)
	fmt.Println(err)

	tx := bst.Heap.Get(tslab.NewP[bst.Node](0, 1))
	fmt.Println(tx, tx.ValueOf, tx.ValueOf.LH, tx.ValueOf.RH)

	bs := strings.Builder{}
	bst.Print(&bs, 0, tx.ValueOf)
	fmt.Println(bs.String())
}

// b, err := json.Marshal(tree.Unpack())
// fmt.Printf("%v %v\n", err, string(b))

// var nn bst.Node
// err = json.Unmarshal(b, &nn)
// fmt.Printf("%v %v\n", err, nn)

// fmt.Printf("binary tree\t%v nodes\n", n)
// fmt.Printf("created in\t%v\t(%v ns/op)\n", tc, tc.Nanoseconds()/int64(n))
// fmt.Printf("folded in\t%v\t(%v ns/op)\n", tf, tf.Nanoseconds()/int64(n))

// runtime.GC()
// MemUsage()

func MemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %7v B", m.Alloc)
	fmt.Printf("\tTotalAlloc = %7v B", m.TotalAlloc)
	fmt.Printf("\tSys = %7v B", m.Sys)
	fmt.Printf("\tNumGC = %v", m.NumGC)
	fmt.Println()
}
