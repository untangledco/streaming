package cair

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"os"
	"strconv"
	"time"
)

type Playlist struct {
	XMLName struct{} `xml:"List"`
	Items   []Item   `xml:"Item"`
}

type Item struct {
	ID            string        `xml:"Id,attr"`
	Name          string        `xml:",attr"`
	Description   string        `xml:",attr"`
	ThirdPartyId  string        `xml:",attr"`
	SubtitleId    string        `xml:",attr"`
	EPGID         string        `xml:"EpgId,attr"`
	ProxyProgress string        `xml:",attr"`
	ScheduledAt   time.Time     `xml:",attr"`
	Duration      time.Duration `xml:",attr"`
	OutOfNetwork  string        `xml:",attr"`
}

// EndTime calculates the time i should end by inspecting its start time and duration.
func (i *Item) EndTime() (time.Time, error) {
	return i.ScheduledAt.Add(i.Duration), nil
}

func (i *Item) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	type alias Item
	aux := &struct {
		ScheduledAt string `xml:",attr"`
		Duration    string `xml:",attr"`
		*alias
	}{
		alias: (*alias)(i),
	}
	if err := dec.DecodeElement(aux, &start); err != nil {
		return err
	}

	t, err := time.Parse(rfc3339Milli, aux.ScheduledAt)
	if err != nil {
		return fmt.Errorf("parse scheduled at: %w", err)
	}
	i.ScheduledAt = t
	dur, err := parseDuration(aux.Duration)
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}
	i.Duration = dur
	return nil
}

const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

// parseDuration parses a time.Duration from timecode string.
// A timecode represents a duration of time.
// A duration of 1 second is represented by the timecode "00:00:01.000".
// 12 hours, 34 minutes, 56 seconds and 789 milliseconds is represented by the timecode
// "12:34:56.789".
func parseDuration(timecode string) (time.Duration, error) {
	if len(timecode) != 12 {
		return 0, fmt.Errorf("timecode does not have 12 characters")
	}

	var duration time.Duration
	hours, err := strconv.ParseInt(string(timecode[:2]), 10, 0)
	if err != nil {
		return 0, fmt.Errorf("parse hours: %v", err)
	}
	duration += time.Duration(hours) * time.Hour

	if string(timecode[2]) != ":" {
		return 0, fmt.Errorf("parse minutes: expected %q, got %q", ":", timecode[3])
	}
	minutes, err := strconv.ParseInt(string(timecode[3:5]), 10, 0)
	if err != nil {
		return 0, fmt.Errorf("parse minutes: %v", err)
	}
	duration += time.Duration(minutes) * time.Minute

	if string(timecode[5]) != ":" {
		return 0, fmt.Errorf("parse seconds: expected %q, got %q", ":", timecode[5])
	}
	seconds, err := strconv.ParseInt(string(timecode[6:8]), 10, 0)
	if err != nil {
		return 0, fmt.Errorf("parse seconds: %v", err)
	}
	duration += time.Duration(seconds) * time.Second

	if string(timecode[8]) != "." {
		return 0, fmt.Errorf("parse milliseconds: expected %q, got %q", ".", timecode[6])
	}
	milliseconds, err := strconv.ParseInt(string(timecode[9:]), 10, 0)
	if err != nil {
		return 0, fmt.Errorf("parse milliseconds: %v", err)
	}
	duration += time.Duration(milliseconds) * time.Millisecond

	return duration, nil
}

func playlistFromFile(name string) (*Playlist, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParsePlaylist(f)
}

// ParsePlaylist parses a XML-encoded Playlist from r.
func ParsePlaylist(r io.Reader) (*Playlist, error) {
	var p Playlist
	escaped := escapeAmpersand(r)
	if err := xml.NewDecoder(escaped).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// escapeAmpersand returns a new reader which reads from r, escaping all ampersand
// characters according to the XML spec.
// This is a hack to make Playlists encoded as invalid XML from third party sources valid.
func escapeAmpersand(r io.Reader) io.Reader {
	buf := &bytes.Buffer{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		buf.Write(bytes.ReplaceAll(scanner.Bytes(), []byte("&"), []byte(html.EscapeString("&"))))
	}
	return buf
}
