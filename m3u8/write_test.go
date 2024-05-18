package m3u8

import (
	"bufio"
	"bytes"
	"testing"
	"time"
)

func TestEncodeSegDuration(t *testing.T) {
	plist := &Playlist{
		Version:  7,
		Segments: []Segment{{Duration: 9967 * time.Millisecond}},
	}
	buf := &bytes.Buffer{}
	if err := Encode(buf, plist); err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(buf)
	var linenum = 1
	var found bool
	want := "#EXTINF:9.967"
	for sc.Scan() {
		t.Log(sc.Text())
		if sc.Text() == want {
			found = true
			return
		}
		linenum++
	}
	if !found {
		t.Errorf("no matching segment duration %s", want)
	}
}
