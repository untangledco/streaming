package m3u8

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteSessionData(t *testing.T) {
	var cases = []struct {
		name  string
		sd    SessionData
		want  string
		valid bool
	}{
		{
			"all attributes set",
			SessionData{ID: "1234", Value: "5678", Language: "indonesian"},
			`#EXT-X-SESSION-DATA:DATA-ID="1234",VALUE="5678",LANGUAGE="indonesian"`,
			true,
		},
		{
			"language is optional",
			SessionData{ID: "1234", URI: "hello/hi.json"},
			`#EXT-X-SESSION-DATA:DATA-ID="1234",URI="hello/hi.json"`,
			true,
		},
		{
			"required attributes set(id, value or uri)",
			SessionData{ID: "1234", Value: "5678"},
			`#EXT-X-SESSION-DATA:DATA-ID="1234",VALUE="5678"`,
			true,
		},
		{
			"empty id",
			SessionData{URI: "hello/hi.json"},
			"",
			false,
		},
		{
			"both URI and Value set",
			SessionData{ID: "1234", URI: "hello/hi.json", Value: "5678"},
			"",
			false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			_, err := writeSessionData(buf, tt.sd)
			if err == nil && !tt.valid {
				t.Fatalf("nil error on invalid session data %v", tt.sd)
			} else if err != nil && tt.valid {
				t.Fatalf("non-nil error %v on valid session data", err)
			}
			got := strings.TrimSpace(buf.String()) // strip newlines
			if got != tt.want {
				t.Errorf("unexpected session data text")
				t.Log("got:", got)
				t.Log("want:", tt.want)
			}
		})
	}
}
