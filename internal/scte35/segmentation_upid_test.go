package scte35

import "testing"

func TestSegmentationUPID_ASCIIValue(t *testing.T) {
	upid := SegmentationUPID{Type: 0x09, Value: "SIGNAL:z1sFOMjCnV4AAAAAAAABAQ=="}
	want := "SIGNAL:z1sFOMjCnV4AAAAAAAABAQ=="
	if want != upid.ASCIIValue() {
		t.Errorf("want %s, got %s", want, upid.ASCIIValue())
	}
}
