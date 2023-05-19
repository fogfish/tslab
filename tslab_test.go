//
// Copyright (C) 2023 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/tslab
//

package tslab_test

import (
	"testing"

	"github.com/fogfish/tslab"
)

type T struct {
	Key int
	Val string
}

func BenchmarkAlloc(b *testing.B) {
	heap := tslab.New[T](16 * 1024)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		heap.Alloc()
	}
}

func BenchmarkAllocFree(b *testing.B) {
	heap := tslab.New[T](16 * 1024)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := heap.Alloc()
		heap.Free(p)
	}
}

// func BenchmarkRef(b *testing.B) {
// 	heap := tslab.New[T](16*1024, unsafe.Sizeof(T{}))
// 	p := heap.Alloc()

// 	b.ReportAllocs()
// 	for i := 0; i < b.N; i++ {
// 		heap.Ref(p)
// 	}
// }
