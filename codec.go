package tslab

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

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

// ---------------------------------------------------------------

func (slot slot) MarshalBinary() ([]byte, error) {
	var buf [12]byte

	binary.BigEndian.PutUint32(buf[0:4], uint32(slot.refs))
	binary.BigEndian.PutUint32(buf[4:8], uint32(slot.self.slabID))
	binary.BigEndian.PutUint32(buf[8:12], uint32(slot.self.slotID))

	return buf[:], nil
}

func (slot *slot) UnmarshalBinary(data []byte) error {
	slot.refs = int32(binary.BigEndian.Uint32(data[0:4]))
	slot.self.slabID = int(binary.BigEndian.Uint32(data[4:8]))
	slot.self.slotID = int(binary.BigEndian.Uint32(data[8:12]))
	slot.next = nilRef
	return nil
}

// ---------------------------------------------------------------

const slabCodecVersion = uint32(0x0001)

func (slab tslab[T]) MarshalBinary() ([]byte, error) {
	var wire bytes.Buffer
	enc := gob.NewEncoder(&wire)

	if err := enc.Encode(slabCodecVersion); err != nil {
		return nil, err
	}

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

	var vsn uint32

	if err := dec.Decode(&vsn); err != nil {
		return err
	}

	if vsn != slabCodecVersion {
		return fmt.Errorf("invalid slab version %x, only %x supported", vsn, slabCodecVersion)
	}

	if err := dec.Decode(&p.slots); err != nil {
		return err
	}

	if err := dec.Decode(&p.memory); err != nil {
		return err
	}

	return nil
}

// ---------------------------------------------------------------

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

	for slabID := 0; slabID < len(slabs.seq); slabID++ {
		slabs.remapSlab(slabID)
	}

	return nil
}

// ---------------------------------------------------------------

func (h *Heap[T]) MarshalBinary() ([]byte, error) {
	var wire bytes.Buffer
	enc := gob.NewEncoder(&wire)

	if err := enc.Encode(h.slabs); err != nil {
		return nil, err
	}

	return wire.Bytes(), nil
}

func (h *Heap[T]) UnmarshalBinary(data []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	if err := dec.Decode(&h.slabs); err != nil {
		return err
	}

	return nil
}
