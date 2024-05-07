package scte35

import (
	"fmt"
	"testing"
)

func TestPackTier(t *testing.T) {
	want := maxTier - 1
	fmt.Printf("%012b\n", want)

	packed := packTier(want)
	got := uint16(packed[1]) | uint16(packed[0])<<8
	if got != want {
		t.Errorf("want packed tier %d, got %d", want, got)
	}
}
