package scte35

import (
	"testing"
)

func TestPackBreakDuration(t *testing.T) {
	dur := uint64(8589934492) // 2^33 - 100
	bd := BreakDuration{true, dur}
	want := [5]byte{
		0b10000001,
		0b11111111,
		0b11111111,
		0b11111111,
		0b10011100,
	}
	got := packBreakDuration(&bd)
	if want != got {
		t.Errorf("packBreakDuration(%v) = %08b, want %08b", bd, got, want)
	}
}
