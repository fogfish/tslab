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

// Typed pointer to the object on the heap
type Pointer[T any] struct {
	ValueOf *T
	slabID  int
	slotID  int
}

func (p Pointer[T]) String() string {
	return fmt.Sprintf("(%p | %d, %d)", p.ValueOf, p.slabID, p.slotID)
}

func (p Pointer[T]) IsNil() bool { return p.slabID == 0 && p.slotID == 0 }

// Memory allocator
type Allocator[T any] interface {
	Alloc() Pointer[T]
	Free(Pointer[T])
}

// Object relocator after recovery
type Relocator[T any] interface {
	Reloc(Pointer[T]) Pointer[T]
}

// Heap statistic
type Stats struct {
	NumAllocs int
	NumFree   int
	Slabs     int
	Slots     int
	SlotsFree int
}

// Typed Heap manages pool of objects on typed-slabs
type Heap[T any] struct {
	slabs tslabs[T]

	statsAllocs int
	statsFree   int
}

// Sequence of typed-slabs
type tslabs[T any] struct {
	seq []*tslab[T]

	slabObjects int

	freeSlots      ref
	freeSlotsCount int
	slotsCount     int
}

// typed slab
type tslab[T any] struct {
	status tslabStatus

	// slab memory
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
	tslab_present tslabStatus = iota
	tslab_swapped
	tslab_dirty
)

// Create new heap for objects of type T
func New[T any](objectsPerChunk int) *Heap[T] {
	return &Heap[T]{
		slabs: tslabs[T]{
			slabObjects: objectsPerChunk,
			freeSlots:   nilRef,
		},
	}
}

// Allocate object from heap
func (h *Heap[T]) Alloc() Pointer[T] {
	slot := h.slabs.alloc()
	// slab := h.slabs.seq[slot.self.slabID]
	// if slab.status != tslab_dirty {
	// 	slab.status = tslab_dirty
	// }

	obj := h.slabs.fetchSlot(slot)

	h.statsAllocs++

	return Pointer[T]{
		ValueOf: obj,
		slabID:  slot.self.slabID,
		slotID:  slot.self.slotID,
	}
}

// Rebinds pointer
func (h *Heap[T]) Reloc(p Pointer[T]) Pointer[T] {
	return h.slabs.Reloc(p)
}

// Free memory allocated to pointer
func (h *Heap[T]) Free(p Pointer[T]) {
	ref := ref{p.slabID, p.slotID}
	slot := h.slabs.refToSlot(ref)
	if slot.refs == 0 {
		return
	}

	slot.refs = 0
	h.slabs.enqueueFreeSlot(slot)

	h.statsFree++

	// slab := h.slabs.seq[slot.self.slabID]
	// if slab.status != tslab_dirty {
	// 	slab.status = tslab_dirty
	// 	h.statsSlabsDirty++
	// }

}

func (h *Heap[T]) Stats() Stats {
	return Stats{
		NumAllocs: h.statsAllocs,
		NumFree:   h.statsFree,
		Slabs:     len(h.slabs.seq),
		// SlabsDirty: h.statsSlabsDirty,
		Slots:     h.slabs.slotsCount,
		SlotsFree: h.slabs.freeSlotsCount,
	}
}

func (h *Heap[T]) Dump() {
	for _, x := range h.slabs.seq {
		fmt.Printf("slab: %v\n", x.status)
		fmt.Printf("%+v\n", x.memory)
	}
}

// func (h *Heap[T]) DumpX() ([]byte, error) {
// 	var wire bytes.Buffer
// 	enc := gob.NewEncoder(&wire)

// 	if err := enc.Encode(h.slabs); err != nil {
// 		return nil, err
// 	}

// 	return wire.Bytes(), nil
// }

// func (h *Heap[T]) UnDump(data []byte) error {
// 	dec := gob.NewDecoder(bytes.NewBuffer(data))

// 	if err := dec.Decode(&h.slabs); err != nil {
// 		return err
// 	}

// 	return nil
// }

// ---------------------------------------------------------------

// allocates slab and slot from heap
func (slabs *tslabs[T]) alloc() *slot {
	if slabs.freeSlots.isNil() {
		slab := slabs.addSlab(len(slabs.seq))
		slabs.seq = append(slabs.seq, slab)
		slabs.slotsCount += slabs.slabObjects
	}

	return slabs.dequeueFreeSlot()
}

// add empty slab
func (slabs *tslabs[T]) addSlab(slabID int) *tslab[T] {
	slab := &tslab[T]{
		status: tslab_present,
		memory: make([]T, slabs.slabObjects, slabs.slabObjects),
		slots:  make([]slot, slabs.slabObjects, slabs.slabObjects),
	}

	for slotID := len(slab.slots); slotID > 0; slotID-- {
		c := &(slab.slots[slotID-1])
		c.self.slabID = slabID
		c.self.slotID = slotID
		slabs.enqueueFreeSlot(c)
	}

	return slab
}

// fetches object and its address behind memory slot
func (slabs *tslabs[T]) fetchSlot(slot *slot) *T {
	slab := slabs.seq[slot.self.slabID]
	slab.status = tslab_dirty

	return &slab.memory[slot.self.slotID-1]
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

// rebind pointers on slab
func (slabs *tslabs[T]) remapSlab(slabID int) {
	slab := slabs.seq[slabID]

	// rebinds pointers for allocated objects
	for slotID := 0; slotID < len(slab.memory); slotID++ {
		obj := &slab.memory[slotID]
		switch vv := any(obj).(type) {
		case interface{ Reloc(Relocator[T]) }:
			vv.Reloc(slabs)
		default:
			fmt.Printf("%T\n", vv)
		}
	}

	// enqueue free slots to heap
	for _, slot := range slab.slots {
		if slot.refs == 0 {
			slabs.enqueueFreeSlot(&slot)
		}
	}
}

// Rebinds pointer
func (slabs *tslabs[T]) Reloc(p Pointer[T]) Pointer[T] {
	if p.slabID == 0 && p.slotID == 0 {
		return p
	}

	ref := ref{slabID: p.slabID, slotID: p.slotID}

	slot := slabs.refToSlot(ref)
	obj := slabs.fetchSlot(slot)

	return Pointer[T]{
		ValueOf: obj,
		slabID:  p.slabID,
		slotID:  p.slotID,
	}
}
