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

func TestReadBreakDuration(t *testing.T) {
	want := samples[1].want.Command.Insert.Duration.Duration
	a := [5]byte{0xfe, 0x00, 0x52, 0xcc, 0xf5}
	got := readBreakDuration(a)
	if want != got.Duration {
		t.Errorf("readBreakDuration(%#x) = %d, want %d", a, got.Duration, want)
	}
}
