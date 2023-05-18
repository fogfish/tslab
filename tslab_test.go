//
// Copyright (C) 2023 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/tslab
//

package tslab_test

import (
	"sync"
	"testing"
	"unsafe"

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
		p, _ := heap.Alloc()
		heap.Free(p)
	}
}

func BenchmarkGet(b *testing.B) {
	heap := tslab.New[T](16 * 1024)
	p, _ := heap.Alloc()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		heap.Get(p)
	}
}

var t *T

func BenchmarkSliceGet(b *testing.B) {
	heap := make([]T, 16*1024, 16*1024)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		t = &heap[i%16*1024]
	}
}

func BenchmarkSliceUnsafe(b *testing.B) {
	heap := make([]byte, 16*1024, 16*1024)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		t = (*T)(unsafe.Pointer(&heap[1024]))
	}
}

func BenchmarkSyncPool(b *testing.B) {
	pool := sync.Pool{}
	pool.New = func() any { return T{} }

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj := pool.Get()
		pool.Put(obj)
	}
}
