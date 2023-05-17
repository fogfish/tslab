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
	"runtime"
	"time"

	"github.com/fogfish/tslab/example/bst"
)

func main() {
	t := time.Now()
	_, tree := bst.New(0, 17)
	tc := time.Since(t)

	// 2^(h+1) - 1
	t = time.Now()
	n := bst.Count(tree)
	tf := time.Since(t)

	fmt.Printf("binary tree\t%v nodes\n", n)
	fmt.Printf("created in\t%v\t(%v ns/op)\n", tc, tc.Nanoseconds()/int64(n))
	fmt.Printf("folded in\t%v\t(%v ns/op)\n", tf, tf.Nanoseconds()/int64(n))

	runtime.GC()
	MemUsage()
}

func MemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %7v B", m.Alloc)
	fmt.Printf("\tTotalAlloc = %7v B", m.TotalAlloc)
	fmt.Printf("\tSys = %7v B", m.Sys)
	fmt.Printf("\tNumGC = %v", m.NumGC)
	fmt.Println()
}
