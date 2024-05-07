package scte35

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
	// next 6 bits are reserved.
	pts := toPTS(b.Duration)
	// 1 bit remaining in the first byte, so pack 1 bit from the timestamp
	p[0] |= pts[0]

	p[1] = pts[1]
	p[2] = pts[2]
	p[3] = pts[3]
	p[4] = pts[4]
	return p
}
