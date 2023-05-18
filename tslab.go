//
// Copyright (C) 2023 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/tslab
//

package tslab

import (
	"fmt"
)

type X[T any] struct {
	Ref *T
	len int
	cap int
}

// typed pointer to the object on Heap
type Pointer[T any] uint64

func (p Pointer[T]) IsNil() bool { return p == 0 }

func (p Pointer[T]) slabID() int { return int(p >> 32) }
func (p Pointer[T]) slotID() int { return int(p & 0xffffffff) }

func newPointer[T any](slabID, slotID int) Pointer[T] {
	return Pointer[T](slabID<<32 | slotID)
}

type Stats struct {
	NumAllocs  int
	Slabs      int
	SlabsDirty int
	Slots      int
	SlotsFree  int
}

// Typed Heap manages pool of objects on typed-slabs
type Heap[T any] struct {
	slabs tslabs[T]

	chunkSize int

	statsAllocs     int
	statsSlabsDirty int
}

// Sequence of typed-slabs
type tslabs[T any] struct {
	seq []*tslab[T]

	freeSlots      Pointer[T]
	freeSlotsCount int
	slotsCount     int
}

// typed slab
type tslab[T any] struct {
	status tslabStatus

	// object heap memory
	memory []T
	// metadata about memory chunks
	slots []slot[T]
}

// metadata about slots available on slab
type slot[T any] struct {
	refs int32
	self Pointer[T]
	next Pointer[T]
}

// status of slab
type tslabStatus int

const (
	tslab_synced tslabStatus = iota
	tslab_dirty
	tslab_swapped
)

// Create new heap for objects of type T
func New[T any](chunkSize int) *Heap[T] {
	return &Heap[T]{
		slabs: tslabs[T]{
			freeSlots: Pointer[T](0),
		},
		chunkSize: chunkSize,
	}
}

// Allocate object from heap
func (h *Heap[T]) Alloc() (Pointer[T], *T) {
	h.statsAllocs++

	slot := h.slabs.alloc(h.chunkSize)
	obj := h.slabs.fetchObject(slot)

	// TODO: slab stats
	// slab := h.slabs.seq[slot.self.slabID]
	// if slab.status != tslab_dirty {
	// 	slab.status = tslab_dirty
	// 	h.statsSlabsDirty++
	// }

	return slot.self, obj
}

// Get object by pointer
func (h *Heap[T]) Get(p Pointer[T]) *T {
	if p.IsNil() {
		return nil
	}

	slot := h.slabs.fetchSlot(p)
	if slot.refs == 0 {
		return nil
	}

	obj := h.slabs.fetchObject(slot)
	return obj
}

// Free memory allocated to pointer
func (h *Heap[T]) Free(p Pointer[T]) {
	if p.IsNil() {
		return
	}

	slot := h.slabs.fetchSlot(p)
	if slot.refs == 0 {
		return
	}

	slab := h.slabs.seq[slot.self.slabID()]
	slot.refs = 0

	if slab.status != tslab_dirty {
		slab.status = tslab_dirty
		h.statsSlabsDirty++
	}

	h.slabs.enqueueFreeSlot(slot)
}

func (h *Heap[T]) Stats() Stats {
	return Stats{
		NumAllocs:  h.statsAllocs,
		Slabs:      len(h.slabs.seq),
		SlabsDirty: h.statsSlabsDirty,
		Slots:      h.slabs.slotsCount,
		SlotsFree:  h.slabs.freeSlotsCount,
	}
}

func (h *Heap[T]) Dump() {
	for _, x := range h.slabs.seq {
		fmt.Printf("slab: %v\n", x.status)
		fmt.Printf("%+v\n", x.memory)
	}
}

// ---------------------------------------------------------------

// allocates slab and slot from heap
func (slabs *tslabs[T]) alloc(size int) *slot[T] {
	if slabs.freeSlots.IsNil() {
		slab := slabs.addSlab(len(slabs.seq), size)
		slabs.seq = append(slabs.seq, slab)
		slabs.slotsCount += size
	}

	return slabs.dequeueFreeSlot()
}

// add empty slab
func (slabs *tslabs[T]) addSlab(slabID, size int) *tslab[T] {
	slab := &tslab[T]{
		status: tslab_synced,
		memory: make([]T, size, size),
		slots:  make([]slot[T], size, size),
	}

	for slotID := len(slab.slots); slotID > 0; slotID-- {
		c := &(slab.slots[slotID-1])
		c.self = newPointer[T](slabID, slotID)
		slabs.enqueueFreeSlot(c)
	}

	return slab
}

// fetches object and its address behind memory slot
func (slabs *tslabs[T]) fetchObject(slot *slot[T]) *T {
	obj := &(slabs.seq[slot.self.slabID()].memory[slot.self.slotID()-1])
	return obj
}

// cast ref to
func (slabs *tslabs[T]) fetchSlot(p Pointer[T]) *slot[T] {
	slot := &(slabs.seq[p.slabID()].slots[p.slotID()-1])
	return slot
}

// release slot to heap
func (slabs *tslabs[T]) enqueueFreeSlot(c *slot[T]) {
	if c.refs != 0 {
		panic("slabs: enqueue no empty slot")
	}

	c.next = slabs.freeSlots
	slabs.freeSlots = c.self
	slabs.freeSlotsCount++
}

// allocate slot from heap
func (slabs *tslabs[T]) dequeueFreeSlot() *slot[T] {
	if slabs.freeSlots.IsNil() {
		panic("slabs: out of memory")
	}

	c := slabs.fetchSlot(slabs.freeSlots)

	if c.refs != 0 {
		panic("slabs: dequeue allocated slot")
	}

	c.refs = 1
	slabs.freeSlots = c.next
	c.next = Pointer[T](0)
	slabs.freeSlotsCount--

	if slabs.freeSlotsCount < 0 {
		panic("slabs: queue of slots corrupted")
	}

	return c
}
