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

func TestBadData(t *testing.T) {
	m := map[string][]byte{
		"garbage":         []byte("0123456789"),
		"only sync bytes": bytes.Repeat([]byte{Sync}, PacketSize),
	}
	for name, in := range m {
		t.Run(name, func(t *testing.T) {
			var p Packet
			if err := Unmarshal(in, &p); err == nil {
				t.Errorf("Unmarshal(%s): nil error", string(in))
			}
		})
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
	var tests = []struct {
		name    string
		encoded [6]byte
		want    PCR
	}{
		{
			"zero",
			[6]byte{0, 0, 0, 0, 0b01111110, 0}, // 6 reserved bits toggled
			PCR{},
		},
		{
			"max base", // 2^33 - 1
			[6]byte{0xff, 0xff, 0xff, 0xff, 0xfe, 0x00},
			PCR{8589934591, 0},
		},
		{
			"max extension", // 2^9 - 1
			[6]byte{0, 0, 0, 0, 0x7f, 0xff},
			PCR{0, 511},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pcr := parsePCR(tt.encoded)
			if pcr != tt.want {
				t.Errorf("parsePCR(%v) = %v, want %v", tt.encoded, pcr, tt.want)
			}

			var a [6]byte
			if err := putPCR(a[:], &pcr); err != nil {
				t.Fatalf("put PCR: %v", err)
			}
			if a != tt.encoded {
				t.Errorf("re-encoded pcr is %v, want %v", a, tt.encoded)
			}
		})
	}
}
