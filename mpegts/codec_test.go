package mpegts

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	f, err := os.Open("testdata/193039199_mp4_h264_aac_hq_7.ts")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for {
		p, err := Decode(f)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatalf("decode packet: %v", err)
		}
		if err := Encode(io.Discard, p); err != nil {
			t.Fatalf("encode packet: %v", err)
		}
	}
}
