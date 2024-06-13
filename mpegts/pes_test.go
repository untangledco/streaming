package mpegts

import (
	"fmt"
	"testing"
)

func TestPackTimestamp(t *testing.T) {
	var tests = []struct {
		tstamp Timestamp
		want   [5]byte
	}{
		{
			Timestamp{true, true, maxTicks},
			[5]byte{0b00111111, 0xff, 0xff, 0xff, 0xff},
		}, {
			Timestamp{true, true, 0},
			[5]byte{0b00110001, 0, 0b00000001, 0, 0b00000001},
		}, {
			// from packet 10 of testdata/193039199_mp4_h264_aac_hq_7.ts
			Timestamp{true, false, 900909},
			[5]byte{0x21, 0x00, 0x37, 0x7e, 0x5b},
		}, {
			// from packet 163 of testdata/193039199_mp4_h264_aac_hq_7.ts
			Timestamp{true, false, 934346},
			[5]byte{0x21, 0x00, 0x39, 0x83, 0x95},
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%d", tt.tstamp.Ticks)
		t.Run(name, func(t *testing.T) {
			packed := packTimestamp(tt.tstamp)
			if packed != tt.want {
				t.Errorf("packTimestamp(%v) = %x, want %x", tt.tstamp, packed, tt.want)
				for i := range packed {
					if packed[i] != tt.want[i] {
						t.Errorf("byte %d: got %08b, want %08b", i, packed[i], tt.want[i])
					}
				}
			}
		})
	}
}
