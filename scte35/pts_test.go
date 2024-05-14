package scte35

import (
	"testing"
)

func TestPTS(t *testing.T) {
	ticks := uint64(8589934591) // max 33-bit uint
	buf := make([]byte, 5)
	putPTS(buf, ticks)
	want := [5]byte{
		0x01,
		0xff,
		0xff,
		0xff,
		0xff,
	}
	var got [5]byte
	copy(got[:], buf)
	if want != got {
		t.Errorf("putPTS(buf, %d); want %#x, got %#x", ticks, want, got)
	}
}
