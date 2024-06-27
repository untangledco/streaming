package cair

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	_, err := playlistFromFile("testdata/playlist.xml")
	if err != nil {
		t.Fatalf("parse playlist %s: %v", "testdata/playlist.xml", err)
	}
}

func TestTimecode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Duration
	}{
		{"zero", "00:00:00.000", 0},
		{"1min15sec", "00:01:15.000", time.Minute + 15*time.Second},
		{
			"longest",
			"12:34:56.789",
			12*time.Hour + 34*time.Minute + 56*time.Second + 789*time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.in)
			if err != nil {
				t.Errorf("parse duration: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v; want %v", got, tt.want)
			}
		})
	}
}

func TestBadTimecode(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"garbage", "世界 Hello"},
		{"decimal", "00:12:009999"},
		{"colon", "00:1122.9999"},
		{"letters", "00:ab:00.000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDuration(tt.in)
			if err == nil {
				t.Errorf("parsing %q succeeded", tt.in)
			}
			t.Log(err)
		})
	}
}
