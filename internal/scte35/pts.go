package scte35

// PTS represents a presentation timestamp - the number of ticks of a
// 90KHz clock - as a 33-bit field.
type PTS [5]byte

func toPTS(ticks uint64) PTS {
	var p PTS
	p[0] = byte(ticks >> 32)
	// mask off 7 bits; we only want 33 total, not 40.
	p[0] &= 0b00000001
	p[1] = byte(ticks >> 24)
	p[2] = byte(ticks >> 16)
	p[3] = byte(ticks >> 8)
	p[4] = byte(ticks)
	return p
}

func ticks(pts PTS) uint64 {
	return uint64(pts[4]) | uint64(pts[3])<<8 | uint64(pts[2])<<16 | uint64(pts[1])<<24 | uint64(pts[0])<<32
}
