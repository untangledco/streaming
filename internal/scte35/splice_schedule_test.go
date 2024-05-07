package scte35

import (
	"testing"
	"time"
)

func TestPackEvent(t *testing.T) {
	ev := Event{
		ID:            6969,
		SpliceTime:    time.Now().Add(5 * time.Second),
		ProgramID:     uint16(500),
		AvailNum:      4,
		AvailExpected: 2,
	}
	packEvent(&ev)
}
