package m3u8

import (
	"bufio"
	"bytes"
	"encoding/base64"
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

func TestDateRange(t *testing.T) {
	// splice insert from SCTE 35 section 14.2
	out, err := base64.StdEncoding.DecodeString("/DAvAAAAAAAA///wFAVIAACPf+/+c2nALv4AUsz1AAAAAAAKAAhDVUVJAAABNWLbowo=")
	if err != nil {
		t.Fatal(err)
	}
	dr := DateRange{
		ID:     "break",
		CueOut: out,
	}
	buf := &bytes.Buffer{}
	if err := writeDateRange(buf, &dr); err != nil {
		t.Fatalf("write date range: %v", err)
	}
	want := `#EXT-X-DATERANGE:ID="break",SCTE35-OUT=0xfc302f000000000000fffff014054800008f7feffe7369c02efe0052ccf500000000000a0008435545490000013562dba30a` + "\n"
	if buf.String() != want {
		t.Errorf("encode %v: got %s, want %s", dr, buf.String(), want)
	}
}
