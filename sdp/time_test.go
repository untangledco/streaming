package sdp

import (
	"reflect"
	"testing"
	"time"
)

func TestParseTimes(t *testing.T) {
	var cases = []struct {
		when string
		want [2]time.Time
	}{
		{"0 0", [2]time.Time{time.Time{}, time.Time{}}},
		{
			"3930082049 0",
			[2]time.Time{
				time.Date(2024, time.July, 16, 1, 27, 29, 0, time.UTC),
				time.Time{},
			},
		},
		{
			"3724394400 3724398000",
			[2]time.Time{
				time.Date(2018, time.January, 8, 10, 0, 0, 0, time.UTC),
				time.Date(2018, time.January, 8, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			"3724484400 3724488000",
			[2]time.Time{
				time.Date(2018, time.January, 9, 11, 0, 0, 0, time.UTC),
				time.Date(2018, time.January, 9, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range cases {
		got, err := parseTimes(tt.when)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("parseTimes(%q) = %v, want %v", tt.when, got, tt.want)
		}
	}
}

// TODO(otl): tests for invalid repeat lines, e.g. missing fields, negative values

func TestParseRepeat(t *testing.T) {
	line := "604800 3600 0 90000"
	want := Repeat{7 * 24 * time.Hour, time.Hour, []time.Duration{0, 25 * time.Hour}}
	got, err := parseRepeat(line)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parse repeat %s: got %v, want %v", line, got, want)
	}
}

func TestDuration(t *testing.T) {
	var cases = []struct{
		s string
		want time.Duration
	}{
		{"2d", 48*time.Hour},
		{"-1d", -24*time.Hour},
	}
	for _, tt := range cases {
		got, err := parseDuration(tt.s)
		if err != nil {
			t.Fatalf("parse %q: %v", tt.s, err)
		}
		if got != tt.want {
			t.Errorf("parseDuration(%q) = %s, want %s", tt.s, got, tt.want)
		}
	}
}

func TestParseAdjustments(t *testing.T) {
	var cases = []struct {
		name    string
		line    string
		want    []TimeAdjustment
		wantErr bool
	}{
		{"uneven", "3730922900 -2h 123456789", nil, true},
		{"missing offset", "3730922900 ", nil, true},
		{"garbage", "hello world!", nil, true},
		{
			"from rfc",
			"3730928400 -1h 3749680800 0",
			[]TimeAdjustment{
				{time.Date(2018, time.March, 25, 1, 0, 0, 0, time.UTC), -time.Hour},
				{time.Date(2018, time.October, 28, 2, 0, 0, 0, time.UTC), 0},
			},
			false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAdjustments(tt.line)
			if err != nil && !tt.wantErr {
				t.Fatal(err)
			} else if err == nil && tt.wantErr {
				t.Error("unexpected nil error")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAdjustments(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}
