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
			t.Logf("%+v", p)
			if p.Adaptation != nil {
				t.Logf("%+v", p.Adaptation)
			}
			if p.PES != nil {
				t.Logf("%+v", p.PES)
				if p.PES.Header != nil {
					t.Logf("PES header: %+v", p.PES.Header)
				}
			}
			t.Fatalf("encode packet %d: %v", i, err)
		}
		var out [188]byte
		copy(out[:], buf.Bytes())

		if in != out {
			t.Errorf("packet %d: encoded and source bytes differ", i)
			t.Logf("%+v", p)
			if p.Adaptation != nil {
				t.Logf("%+v", p.Adaptation)
			}
			if p.PES != nil {
				t.Logf("%+v", p.PES)
				if p.PES.Header != nil {
					t.Logf("PES header: %+v", p.PES.Header)
				}
			}
			for i := range in {
				if in[i] != out[i] {
					t.Errorf("byte %d: source %08b, encoded %08b", i, in[i], out[i])
					t.Errorf("byte %d: source %#x, encoded %#x", i, in[i], out[i])
				}
			}
		}
	}
}

func TestOnePacket(t *testing.T) {
	tstamp := Timestamp{
		PTS:   true,
		Ticks: 900909,
	}
	var want = [5]byte{0x21, 0x00, 0x37, 0x7e, 0x5b}

	unpacked, err := unpackTimestamp(want)
	if err != nil {
		t.Fatal(err)
	}
	if unpacked.Ticks != tstamp.Ticks {
		t.Errorf("unpacked ticks %d, want %d", unpacked.Ticks, tstamp.Ticks)
		t.Logf("want\t%#033b", tstamp.Ticks)
		t.Logf("got\t%#033b", unpacked.Ticks)
	}

	packed := packTimestamp(tstamp)
	if packed != want {
		t.Errorf("packTimestamp(%v) = %#x, want %#x", tstamp, packed, want)
		t.Logf("got\t%08b", packed)
		t.Logf("want\t%08b", want)
	}
}
