package sdp

import (
	"net/mail"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestReadSession(t *testing.T) {
	var cases = []struct {
		name string
		want Session
		err  bool
	}{
		{
			name: "good.sdp",
			want: Session{
				Name:   "Call to John Smith",
				Origin: Origin{"jdoe", 3724394400, 3724394405, "IP4", "198.51.100.1"},
				Info:   "SDP Offer #1",
				URI: &url.URL{
					Scheme: "http",
					Host:   "www.jdoe.example.com",
					Path:   "/home.html",
				},
				Email: &mail.Address{"Jane Doe", "jane@jdoe.example.com"},
				Phone: "+16175556011",
				Connection: &ConnInfo{
					Type:    "IP4",
					Address: "198.51.100.1",
				},
				Media: []Media{
					Media{
						Type:     "audio",
						Port:     49170,
						Protocol: ProtoRTP,
						Format:   []string{"0"},
					},
					Media{
						Type:     "audio",
						Port:     49180,
						Protocol: ProtoRTP,
						Format:   []string{"0"},
					},
					Media{
						Type:       "video",
						Port:       51372,
						Protocol:   ProtoRTP,
						Format:     []string{"99"},
						Connection: &ConnInfo{"IP6", "2001:db8::2", 0, 0},
						Attributes: []string{"rtpmap:99", "h263-1998/90000"},
					},
				},
			},
		},
		{
			name: "some_optional.sdp",
			want: Session{
				Origin: Origin{"jdoe", 3724394400, 3724394405, "IP4", "198.51.100.1"},
				Name:   "Call to John Smith",
				Email:  &mail.Address{"Jane Doe", "jane@jdoe.example.com"},
			},
		},
		{
			name: "missing_origin.sdp",
			want: Session{},
			err:  true,
		},
		{
			name: "out_of_order.sdp",
			want: Session{},
			err:  true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open("testdata/" + tt.name)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			session, err := ReadSession(f)
			if err == nil && tt.err {
				t.Fatal("unexpected nil error")
			} else if err != nil && tt.err {
				return
			} else if err != nil && !tt.err {
				t.Fatalf("read session: %v", err)
			}
			if !reflect.DeepEqual(*session, tt.want) {
				t.Errorf("got %+v\nwant %+v\n", *session, tt.want)
			}
		})
	}
}

func TestBandwidth(t *testing.T) {
	var cases = []struct {
		name    string
		s       string
		wantErr bool
	}{
		{"conference total", "CT:2048", false},
		{"app specific", "AS:87654321", false},
		{"custom", "69something:12345", false},
		{"missing modifier", ":12345", true},
		{"missing separator", "CT2048", true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseBandwidth(tt.s)
			if err != nil && tt.wantErr {
				// no worries, we got what we expected
			} else if err != nil && !tt.wantErr {
				t.Errorf("parseBandwidth(%q): unexpected error %v", tt.s, err)
			} else if err == nil && tt.wantErr {
				t.Errorf("parseBandwidth(%q): unexpected nil error", tt.s)
			}
		})
	}
}

func TestConnInfo(t *testing.T) {
	var cases = []struct {
		name string
		line string
		want ConnInfo
	}{
		{"ipv4", "IN IP4 192.0.2.1", ConnInfo{"IP4", "192.0.2.1", 0, 0}},
		{"ipv4 ttl", "IN IP4 233.252.0.1/127", ConnInfo{"IP4", "233.252.0.1", 127, 0}},
		{"ipv4 ttl count", "IN IP4 233.252.0.1/127/3", ConnInfo{"IP4", "233.252.0.1", 127, 3}},
		{"ipv6", "IN IP6 2001:db8::1", ConnInfo{"IP6", "2001:db8::1", 0, 0}},
		{"ipv6 count", "IN IP6 ff00::db8:0:101/3", ConnInfo{"IP6", "ff00::db8:0:101", 0, 3}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConnInfo(tt.line)
			if err != nil {
				t.Fatalf("parse %s: %v", tt.line, err)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("parseConnInfo(%q) = %+v, want %+v", tt.line, got, tt.want)
			}
		})
	}
}

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
		t.Run(tt.when, func(t *testing.T) {
			got, err := parseTimes(tt.when)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("parseTimes(%q) = %v, want %v", tt.when, got, tt.want)
			}
		})
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

// TODO(otl): tests for invalid duration strings e.g. bad suffix, no numbers

func TestDuration(t *testing.T) {
	var cases = []struct {
		name    string
		s       string
		want    time.Duration
		wantErr bool
	}{
		{
			name: "dayOfSeconds",
			s:    "86400",
			want: 24 * time.Hour,
		},
		{
			name: "twentyFourHours",
			s:    "24h",
			want: 24 * time.Hour,
		},
		{
			name: "oneDay",
			s:    "1d",
			want: 24 * time.Hour,
		},
		{
			name: "nice",
			s:    "69s",
			want: 69 * time.Second,
		},
		{
			name: "negative",
			s:    "-01s",
			want: time.Duration(-1) * time.Second,
		},
		{
			name: "decimal",
			s:    "1.5h",
			want: time.Duration(5400) * time.Second,
		},
		{
			name:    "badSuffix",
			s:       "13k",
			want:    0,
			wantErr: true,
		},
		{
			name: "2Days",
			s:    "2d",
			want: 48 * time.Hour,
		},
		{
			name:    "aDay",
			s:       "Ad",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.s, func(t *testing.T) {
			got, err := parseDuration(tt.s)
			if (err != nil) != tt.wantErr {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %s, want %s", tt.s, got, tt.want)
			}
		})
	}
}
