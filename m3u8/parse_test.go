package m3u82

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
