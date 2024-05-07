package scte35

import (
	"testing"
)

func TestPTS(t *testing.T) {
	cases := []uint64{1, 128, 8589934492}

	for _, tt := range cases {
		pts := toPTS(tt)
		count := ticks(pts)
		if count != tt {
			t.Errorf("ticks(%b) = %d, want %d", pts, count, tt)
		}
	}
}
