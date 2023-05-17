//
// Copyright (C) 2023 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/tslab
//

package tslab

import "fmt"

// typed pointer to the object on Heap
type Ptr[T any] uint32

func (ptr Ptr[T]) ref() ref {
	return ref{slabID: int(ptr >> 16), slotID: int(ptr & 0xffff)}
}

func (ptr Ptr[T]) IsNil() bool { return ptr == 0 }

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

	freeSlots      ref
	freeSlotsCount int
	slotsCount     int
}

// typed slab
type tslab[T any] struct {
	status tslabStatus

	// object heap memory
	memory []T
	// metadata about memory chunks
	slots []slot
}

// metadata about slots available on slab
type slot struct {
	refs int32
	self ref
	next ref
}

// reference to slot
type ref struct {
	slabID int
	slotID int
}

func (ref ref) isNil() bool { return ref.slabID == -1 && ref.slotID == -1 }

var nilRef = ref{-1, -1}

// status of slab
type tslabStatus int

const (
	tslab_synced tslabStatus = iota
	tslab_dirty
	tslab_swapped
)

// Create new heap for objects of type T
func New[T any](chunkSize int) *Heap[T] {
	if chunkSize >= 64*1024 {
		chunkSize = 64*1024 - 1
	}

	return &Heap[T]{
		slabs:     tslabs[T]{freeSlots: nilRef},
		chunkSize: chunkSize,
	}
}

// Allocate object from heap
func (h *Heap[T]) Alloc() (Ptr[T], *T) {
	h.statsAllocs++

	slot := h.slabs.alloc(h.chunkSize)
	slab := h.slabs.seq[slot.self.slabID]
	if slab.status != tslab_dirty {
		slab.status = tslab_dirty
		h.statsSlabsDirty++
	}

	return h.slabs.fetchSlot(slot)
}

// Get object by pointer
func (h *Heap[T]) Get(ptr Ptr[T]) *T {
	if ptr.IsNil() {
		return nil
	}

	slot := h.slabs.refToSlot(ptr.ref())
	if slot.refs == 0 {
		return nil
	}

	_, obj := h.slabs.fetchSlot(slot)
	return obj
}

// Free memory allocated to pointer
func (h *Heap[T]) Free(ptr Ptr[T]) {
	slot := h.slabs.refToSlot(ptr.ref())
	if slot.refs == 0 {
		return
	}

	slab := h.slabs.seq[slot.self.slabID]
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
func (slabs *tslabs[T]) alloc(size int) *slot {
	if slabs.freeSlots.isNil() {
		slab := slabs.addSlab(len(slabs.seq), size)
		slabs.seq = append(slabs.seq, slab)
		slabs.slotsCount += size
	}

	return slabs.dequeueFreeSlot()
}

// add empty slab
func (slabs *tslabs[T]) addSlab(id, size int) *tslab[T] {
	slab := &tslab[T]{
		status: tslab_synced,
		memory: make([]T, size, size),
		slots:  make([]slot, size, size),
	}

	for i := 1; i <= len(slab.slots); i++ {
		c := &(slab.slots[i-1])
		c.self.slabID = id
		c.self.slotID = i
		slabs.enqueueFreeSlot(c)
	}

	return slab
}

// fetches object and its address behind memory slot
func (slabs *tslabs[T]) fetchSlot(slot *slot) (Ptr[T], *T) {
	slab := slabs.seq[slot.self.slabID]
	slab.status = tslab_dirty

	obj := &(slab.memory[slot.self.slotID-1])
	ptr := Ptr[T](slot.self.slabID<<16 | slot.self.slotID)

	return ptr, obj
}

// cast ref to
func (slabs *tslabs[T]) refToSlot(ref ref) *slot {
	if ref.isNil() {
		return nil
	}

	return &(slabs.seq[ref.slabID].slots[ref.slotID-1])
}

// release slot to heap
func (slabs *tslabs[T]) enqueueFreeSlot(c *slot) {
	if c.refs != 0 {
		panic("slabs: enqueue no empty slot")
	}

	c.next = slabs.freeSlots
	slabs.freeSlots = c.self
	slabs.freeSlotsCount++
}

// allocate slot from heap
func (slabs *tslabs[T]) dequeueFreeSlot() *slot {
	if slabs.freeSlots.isNil() {
		panic("slabs: out of memory")
	}

	c := slabs.refToSlot(slabs.freeSlots)

	if c.refs != 0 {
		panic("slabs: dequeue allocated slot")
	}

	c.refs = 1
	slabs.freeSlots = c.next
	c.next = nilRef
	slabs.freeSlotsCount--

	if slabs.freeSlotsCount < 0 {
		panic("slabs: queue of slots corrupted")
	}

	return c
}
