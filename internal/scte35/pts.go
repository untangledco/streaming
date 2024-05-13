package scte35

// PTS represents a presentation timestamp - the number of ticks of a
// 90KHz clock - as a 33-bit field.
type PTS [5]byte

func toPTS(ticks uint64) [5]byte {
	var p [5]byte
	p[0] = byte(ticks >> 32)
	// mask off 7 bits; we only want 33 total, not 40.
	p[0] &= 0b00000001
	p[1] = byte(ticks >> 24)
	p[2] = byte(ticks >> 16)
	p[3] = byte(ticks >> 8)
	p[4] = byte(ticks)
	return p
}

func putPTS(buf []byte, ticks uint64) {
	pts := toPTS(ticks)
	buf[0] |= pts[0]
	copy(buf[1:5], pts[1:5])
}
