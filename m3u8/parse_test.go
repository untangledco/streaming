package m3u8

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	f, err := os.Open("testdata/bbb.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := ParsePlaylist(f)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(p.Segments[0])
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
