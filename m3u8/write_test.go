package m3u8

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteVariant(t *testing.T) {
	var cases = []struct {
		name  string
		v     Variant
		want  string
		valid bool
	}{
		{
			"simple",
			Variant{
				URI:        "url_0/193039199_mp4_h264_aac_hd_7.m3u8",
				Bandwidth:  2149280,
				Codecs:     []string{"mp4a.40.2", "avc1.64001f"},
				Resolution: [2]int{1280, 720},
			},
			`#EXT-X-STREAM-INF:BANDWIDTH=2149280,CODECS="mp4a.40.2,avc1.64001f",RESOLUTION=1280x720
url_0/193039199_mp4_h264_aac_hd_7.m3u8`,
			true,
		},
		{
			"rounded frame rate",
			Variant{
				URI:       "small.m3u8",
				Bandwidth: 10000,
				FrameRate: 60 / 1.001, // busted NTSC
			},
			`#EXT-X-STREAM-INF:BANDWIDTH=10000,FRAME-RATE=59.940
small.m3u8`,
			true,
		},
		{
			"no bandwidth",
			Variant{URI: "url_0/193039199_mp4_h264_aac_hd_7.m3u8"},
			"",
			false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			_, err := writeVariant(buf, &tt.v)
			if err != nil && tt.valid {
				t.Fatalf("non-nil error for valid variant %v: %v", tt.v, err)
			} else if !tt.valid && err == nil {
				t.Fatalf("nil error for invalid variant %v: %v", tt.v, err)
			}
			got := strings.TrimSpace(buf.String()) // trim newline
			if got != tt.want {
				t.Errorf("unexpected variant text")
				t.Logf("got: %s", got)
				t.Logf("want: %s", tt.want)
			}
		})
	}
}
