package scte35

import (
	"bytes"
	"testing"
)

func TestEncodeInsert(t *testing.T) {
	out := &Insert{
		ID:                12345,
		EventIDCompliance: true,
		OutOfNetwork:      true,
		Immediate:         true,
		Duration:          &BreakDuration{true, uint64(90000 * 2)},
	}
	in := &Insert{
		ID:                out.ID,
		EventIDCompliance: true,
	}

	bout := encodeInsert(out)
	bin := encodeInsert(in)
	if bytes.Equal(bout, bin) {
		t.Errorf("different inserts are equal when encoded")
		t.Log(bout)
		t.Log(bin)
	}
}
