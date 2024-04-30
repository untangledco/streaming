package scte35_test

import (
	"testing"

	"github.com/untangledco/streaming/internal/scte35"
	"github.com/stretchr/testify/require"
)

func TestSegmentationUPID_ASCIIValue(t *testing.T) {
	cases := map[string]struct {
		upid     scte35.SegmentationUPID
		expected string
	}{
		"Simple": {
			upid: scte35.SegmentationUPID{
				Type:  0x09,
				Value: "SIGNAL:z1sFOMjCnV4AAAAAAAABAQ==",
			},
			expected: "SIGNAL:z1sFOMjCnV4AAAAAAAABAQ==",
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			require.Equal(t, c.expected, c.upid.ASCIIValue())
		})
	}
}
