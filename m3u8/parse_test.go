package m3u8

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func ExampleEncode() {
	p := &Playlist{
		Version: 7,
		Segments: []Segment{
			{URI: "001.ts", Duration: 4 * time.Second},
		},

		TargetDuration: 4 * time.Second,
		Sequence:       0,
		Type:           PlaylistEvent,
	}

	sb := &strings.Builder{}
	_ = Encode(sb, p)

	fmt.Println(sb)

	// Output: #EXTM3U
	// #EXT-X-VERSION:7
	// #EXT-X-PLAYLIST-TYPE:EVENT
	// #EXT-X-TARGETDURATION:4
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXTINF:4.000
	// 001.ts
}

func ExampleDecode() {
	s := `
#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=640x360
url_0/low.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1280x720
url_0/mid.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=7680000,RESOLUTION=1920x1080
url_0/high.m3u8`

	p, err := Decode(strings.NewReader(s))
	if err != nil {
		// handle error
	}
	for _, v := range p.Variants {
		fmt.Printf("%s %dp@%dkbps\n", v.URI, v.Resolution[1], v.Bandwidth/1e3)
	}
	// Output:
	// url_0/low.m3u8 360p@1280kbps
	// url_0/mid.m3u8 720p@2560kbps
	// url_0/high.m3u8 1080p@7680kbps
}

func TestDecode(t *testing.T) {
	names, err := filepath.Glob("testdata/*.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		t.Run(path.Base(name), func(t *testing.T) {
			f, err := os.Open(name)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			_, err = Decode(f)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	want := 9967000 * time.Microsecond
	it := item{typ: itemNumber, val: "9.967"}
	dur, err := parseSegmentDuration(it)
	if err != nil {
		t.Fatal(err)
	}
	if dur != want {
		t.Errorf("parseSegmentDuration(%s) = %s, want %s", it, dur, want)
	}
}

func TestParseByteRange(t *testing.T) {
	var tests = []struct {
		in    string
		want  ByteRange
		valid bool
	}{
		{"27@46", ByteRange{27, 46}, true},
		{"69", ByteRange{69}, true},
		{"732@", ByteRange{0, 0}, false},
		{"@", ByteRange{0, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			r, err := parseByteRange(tt.in)
			if err != nil && tt.valid {
				t.Fatalf("parseByteRange(%s): %v", tt.in, err)
			} else if err == nil && !tt.valid {
				t.Fatalf("parseByteRange(%s): nil error on invalid byte range", tt.in)
			}
			if r != tt.want {
				t.Errorf("parseByteRange(%s) = %v, want %v", tt.in, r, tt.want)
			}
		})
	}
}

// Tests that we parse floats and integers of different precisions ok.
func TestFrameRate(t *testing.T) {
	f, err := os.Open("testdata/frame_rate.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	plist, err := Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	rates := []float32{24, 25, 29.97, 30, 23.976, 60}
	for i, v := range plist.Variants {
		if v.FrameRate != rates[i] {
			t.Errorf("want %f, got %f", rates[i], v.FrameRate)
		}
	}
}

func TestParseSequence(t *testing.T) {
	f, err := os.Open("testdata/sequence.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	plist, err := Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if plist.Sequence != 91240 {
		t.Errorf("want %d, got %d", 91240, plist.Sequence)
	}
}
