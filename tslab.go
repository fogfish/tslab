//
// Copyright (C) 2023 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/tslab
//

package tslab

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

type Pointer[T any] struct {
	ValueOf *T
	slabID  int
	slotID  int
}

func (p Pointer[T]) String() string {
	return fmt.Sprintf("(%p | %d, %d)", p.ValueOf, p.slabID, p.slotID)
}

func NewP[T any](slabID, slotID int) Pointer[T] {
	return Pointer[T]{slabID: slabID, slotID: slotID}
}

func (p Pointer[T]) MarshalBinary() ([]byte, error) {
	var buf [8]byte

	binary.BigEndian.PutUint32(buf[0:4], uint32(p.slabID))
	binary.BigEndian.PutUint32(buf[4:8], uint32(p.slotID))

	return buf[:], nil
}

func (p *Pointer[T]) UnmarshalBinary(data []byte) error {
	p.slabID = int(binary.BigEndian.Uint32(data[0:4]))
	p.slotID = int(binary.BigEndian.Uint32(data[4:8]))
	return nil
}

// Heap statistic
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

	statsAllocs     int
	statsSlabsDirty int
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
	refs int32 // 4
	self ref   // 4 + 4
	next ref   // 4 + 4
}

func (slot slot) MarshalBinary() ([]byte, error) {
	var buf [12]byte

	binary.BigEndian.PutUint32(buf[0:4], uint32(slot.refs))
	binary.BigEndian.PutUint32(buf[4:8], uint32(slot.self.slotID))
	binary.BigEndian.PutUint32(buf[8:12], uint32(slot.self.slabID))

	return buf[:], nil
}

func (slot *slot) UnmarshalBinary(data []byte) error {
	// TODO: fix order
	slot.refs = int32(binary.BigEndian.Uint32(data[0:4]))
	slot.self.slotID = int(binary.BigEndian.Uint32(data[4:8]))
	slot.self.slabID = int(binary.BigEndian.Uint32(data[8:12]))
	slot.next = nilRef
	return nil
}

// reference to slot
type ref struct {
	slabID int
	slotID int // TODO: slotAddr
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
func New[T any](slabObjects int) *Heap[T] {
	return &Heap[T]{
		slabs: tslabs[T]{
			slabObjects: slabObjects,
			freeSlots:   nilRef,
		},
	}
}

// Allocate object from heap
func (h *Heap[T]) Alloc() Pointer[T] {
	h.statsAllocs++

	slot := h.slabs.alloc()
	slab := h.slabs.seq[slot.self.slabID]
	if slab.status != tslab_dirty {
		slab.status = tslab_dirty
		h.statsSlabsDirty++
	}

	obj := h.slabs.fetchSlot(slot)

	return Pointer[T]{
		ValueOf: obj,
		slabID:  slot.self.slabID,
		slotID:  slot.self.slotID,
	}
}

// Rebinds pointer
func (h *Heap[T]) Get(p Pointer[T]) Pointer[T] {
	return h.slabs.Get(p)
}

// Free memory allocated to pointer
func (h *Heap[T]) Free(p Pointer[T]) {
	ref := ref{p.slabID, p.slotID}
	slot := h.slabs.refToSlot(ref)
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

func (h *Heap[T]) DumpX() ([]byte, error) {
	var wire bytes.Buffer
	enc := gob.NewEncoder(&wire)

	if err := enc.Encode(h.slabs); err != nil {
		return nil, err
	}

	return wire.Bytes(), nil
}

func (h *Heap[T]) UnDump(data []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	if err := dec.Decode(&h.slabs); err != nil {
		return err
	}

	return nil
}

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
		status: tslab_synced,
		memory: make([]T, slabs.slabObjects, slabs.slabObjects),
		slots:  make([]slot, slabs.slabObjects, slabs.slabObjects),
	}

	for slotID := len(slab.slots) - 1; slotID >= 0; slotID-- {
		c := &(slab.slots[slotID])
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

	return &slab.memory[slot.self.slotID]
}

// cast ref to
func (slabs *tslabs[T]) refToSlot(ref ref) *slot {
	if ref.isNil() {
		return nil
	}

	return &(slabs.seq[ref.slabID].slots[ref.slotID])
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

func (slabs tslabs[T]) MarshalBinary() ([]byte, error) {
	var wire bytes.Buffer
	enc := gob.NewEncoder(&wire)

	if err := enc.Encode(slabs.seq); err != nil {
		return nil, err
	}

	return wire.Bytes(), nil
}

func (slabs *tslabs[T]) UnmarshalBinary(data []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	if err := dec.Decode(&slabs.seq); err != nil {
		return err
	}

	for _, slab := range slabs.seq {
		for _, slot := range slab.slots {
			if slot.refs == 0 {
				slabs.enqueueFreeSlot(&slot)
			}
		}
	}

	fmt.Printf("===> %v\n", len(slabs.seq))

	for slabID := 0; slabID < len(slabs.seq); slabID++ {
		slab := slabs.seq[slabID]
		for slotID := 0; slotID < len(slab.memory); slotID++ {
			obj := &slab.memory[slotID]
			switch vv := any(obj).(type) {
			case interface {
				SwapOut(interface{ Get(Pointer[T]) Pointer[T] })
			}:
				vv.SwapOut(slabs)
			default:
				fmt.Printf("%T\n", vv)
			}
		}

		// fmt.Printf("%+v\n", slab.memory)
	}

	return nil
}

// Rebinds pointer
func (slabs *tslabs[T]) Get(p Pointer[T]) Pointer[T] {
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

// ---------------------------------------------------------------

func (slab tslab[T]) MarshalBinary() ([]byte, error) {
	var wire bytes.Buffer
	enc := gob.NewEncoder(&wire)

	if err := enc.Encode(slab.slots); err != nil {
		return nil, err
	}

	if err := enc.Encode(slab.memory); err != nil {
		return nil, err
	}

	return wire.Bytes(), nil
}

func (p *tslab[T]) UnmarshalBinary(data []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	if err := dec.Decode(&p.slots); err != nil {
		return err
	}

	if err := dec.Decode(&p.memory); err != nil {
		return err
	}

	return nil
}
