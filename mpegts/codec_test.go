package mpegts

import (
	"bytes"
	"crypto/md5"
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

func TestScanner(t *testing.T) {
	name := "testdata/193039199_mp4_h264_aac_hq_7.ts"
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	sum := md5.Sum(data)

	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf := &bytes.Buffer{}
	sc := NewScanner(f)
	var i int
	for sc.Scan() {
		i++
		p := sc.Packet()
		if err := Encode(buf, p); err != nil {
			t.Fatalf("packet %d: encode: %v", i, err)
		}
	}
	if sc.Err() != nil {
		t.Fatalf("scan: %v", sc.Err())
	}
	got := md5.Sum(buf.Bytes())
	if got != sum {
		t.Errorf("re-encoded stream differs from source: got checksum %x, want %x", got, sum)
	}
}

func TestPCR(t *testing.T) {
	a := [6]byte{0x00, 0x24, 0x52, 0xd4, 0x7e, 0x00}
	pcr := parsePCR(a)

	var got [6]byte
	if err := putPCR(got[:], &pcr); err != nil {
		t.Errorf("put pcr: %v", err)
	}
	if got != a {
		t.Errorf("PCR differs after decode, re-encode")
		t.Errorf("putPCR(buf, %v) = %08b, want %08b", pcr, got, a)
	}
}
