package mpegts

import (
	"bytes"
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
	var i int
	for {
		i++
		var in [188]byte
		n, err := f.Read(in[:])
		if errors.Is(err, io.EOF) {
			break
		} else if n != PacketSize {
			t.Fatalf("short read: read %d bytes", n)
		} else if err != nil {
			t.Fatalf("read packet: %v", err)
		}
		p, err := Decode(bytes.NewReader(in[:]))
		if err != nil {
			t.Fatalf("decode packet: %v", err)
		}

		buf := &bytes.Buffer{}
		if err := Encode(buf, p); err != nil {
			t.Fatalf("encode packet %d: %v", i, err)
		}
		var out [188]byte
		copy(out[:], buf.Bytes())

		if in != out {
			t.Errorf("packet %d: encoded and source bytes differ", i)
			t.Logf("%+v", p)
			for i := range in {
				if in[i] != out[i] {
					t.Errorf("byte %d: source %v, encoded %v", i, in[i], out[i])
				}
			}
		}
	}
}
