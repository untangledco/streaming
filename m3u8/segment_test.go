package m3u8

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestMarshalSegments(t *testing.T) {
	var cases = []struct {
		name string
		seg  Segment
		out  string
	}{
		{
			"duration",
			Segment{Duration: 10 * time.Second, URI: "bunny.ts"},
			"#EXTINF:10.000\nbunny.ts",
		},
		{
			"duration milliseconds",
			Segment{URI: "something.ts", Duration: 9967 * time.Millisecond},
			"#EXTINF:9.967\nsomething.ts",
		},
		{
			"discontinuity with URI",
			Segment{
				Duration:      30 * time.Second,
				Discontinuity: true,
				URI:           "adbreak.ts",
			},
			"#EXT-X-DISCONTINUITY\n#EXTINF:30.000\nadbreak.ts",
		},
		{
			"byte range",
			Segment{
				Duration: 2 * time.Second,
				URI:      "vid.ts",
				Range:    ByteRange{69, 420},
			},
			"#EXT-X-BYTERANGE:69@420\n#EXTINF:2.000\nvid.ts",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.seg.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			got := string(b)
			if got != tt.out {
				t.Errorf("segment text does not match expected")
				t.Log("got:", got)
				t.Log("want:", tt.out)
			}
		})
	}
}

func TestMarshalBadSegments(t *testing.T) {
	var cases = []struct {
		name string
		seg  Segment
	}{
		{"empty", Segment{}},
		{"no duration", Segment{URI: "video.ts"}},
		{"impossible range", Segment{URI: "bbb.ts", Duration: 6 * time.Second, Range: ByteRange{999, 10}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.seg.MarshalText(); err == nil {
				t.Fatalf("nil error encoding invalid segment")
			}
		})
	}
}

func TestWriteKey(t *testing.T) {
	var iv [16]byte
	binary.LittleEndian.PutUint64(iv[:8], 10000)
	binary.LittleEndian.PutUint64(iv[8:], 98765432)
	k := Key{
		Method:         EncryptMethodAES128,
		URI:            "magic.key",
		IV:             iv,
		Format:         defaultKeyFormat,
		FormatVersions: []uint32{1, 2, 5},
	}
	want := `#EXT-X-KEY:METHOD=AES-128,URI="magic.key",IV=0x1027000000000000780ae30500000000,KEYFORMAT="identity",KEYFORMATVERSIONS="1/2/5"`
	if k.String() != want {
		t.Errorf("unexpected segment key text")
		t.Log("got:", k.String())
		t.Log("want:", want)
	}
}
