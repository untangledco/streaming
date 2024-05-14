package scte35

import (
	"encoding/binary"
)

type BreakDuration struct {
	AutoReturn bool
	// Holds a number of ticks of a 90KHz clock.
	Duration uint64
}

func packBreakDuration(b *BreakDuration) [5]byte {
	var p [5]byte
	if b.AutoReturn {
		p[0] |= (1 << 7)
	}
	// toggle 6 reserved bits
	p[0] |= 0b01111110
	pts := toPTS(b.Duration)
	// 1 bit remaining in the first byte, so pack 1 bit from the timestamp
	p[0] |= pts[0]

	p[1] = pts[1]
	p[2] = pts[2]
	p[3] = pts[3]
	p[4] = pts[4]
	return p
}

func readBreakDuration(a [5]byte) *BreakDuration {
	var bd BreakDuration
	if a[0]&(1<<7) > 0 {
		bd.AutoReturn = true
	}
	a[0] &= 0x01
	// first, allocate 3 empty bytes, then add the remaining 5;
	// enough for the uint64 (8 bytes).
	b := []byte{0, 0, 0}
	b = append(b, a[:]...)
	bd.Duration = binary.BigEndian.Uint64(b)
	return &bd
}
