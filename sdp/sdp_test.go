package sdp

import (
	"net/mail"
	"net/netip"
	"net/url"
	"os"
	"reflect"
	"testing"
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
				Origin: Origin{"jdoe", 3724394400, 3724394405, netip.MustParseAddr("198.51.100.1")},
				Info:   "SDP Offer #1",
				URI: &url.URL{
					Scheme: "http",
					Host:   "www.jdoe.example.com",
					Path:   "/home.html",
				},
				Email: &mail.Address{
					Name:    "Jane Doe",
					Address: "jane@jdoe.example.com",
				},
				Phone: "+16175556011",
				Connection: &ConnInfo{
					Address: netip.MustParseAddr("198.51.100.1"),
				},
				Media: []Media{
					Media{
						Type:      MediaTypeAudio,
						Port:      49170,
						Transport: ProtoRTP,
						Format:    []string{"0"},
					},
					Media{
						Type:      MediaTypeAudio,
						Port:      49180,
						Transport: ProtoRTP,
						Format:    []string{"0"},
					},
					Media{
						Type:       MediaTypeVideo,
						Port:       51372,
						Transport:  ProtoRTP,
						Format:     []string{"99"},
						Connection: &ConnInfo{netip.MustParseAddr("2001:db8::2"), 0, 0},
						Attributes: []string{"rtpmap:99", "h263-1998/90000"},
					},
				},
			},
		},
		{
			name: "some_optional.sdp",
			want: Session{
				Origin: Origin{"jdoe", 3724394400, 3724394405, netip.MustParseAddr("198.51.100.1")},
				Name:   "Call to John Smith",
				Email: &mail.Address{
					Name:    "Jane Doe",
					Address: "jane@jdoe.example.com",
				},
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
		{"ipv4", "IN IP4 192.0.2.1", ConnInfo{netip.MustParseAddr("192.0.2.1"), 0, 0}},
		{"ipv4 ttl", "IN IP4 233.252.0.1/127", ConnInfo{netip.MustParseAddr("233.252.0.1"), 127, 0}},
		{"ipv4 ttl count", "IN IP4 233.252.0.1/127/3", ConnInfo{netip.MustParseAddr("233.252.0.1"), 127, 3}},
		{"ipv6", "IN IP6 2001:db8::1", ConnInfo{netip.MustParseAddr("2001:db8::1"), 0, 0}},
		{"ipv6 count", "IN IP6 ff00::db8:0:101/3", ConnInfo{netip.MustParseAddr("ff00::db8:0:101"), 0, 3}},
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
