package sip

import (
	"io"
	"net/textproto"
	"os"
	"strings"
	"testing"
)

func TestWriteRequest(t *testing.T) {
	header := make(textproto.MIMEHeader)
	header.Set("Call-ID", "a84b4c76e66710@pc33.example.com")
	header.Set("CSeq", "314159 "+MethodInvite)
	header.Set("Contact", "<sip:alice@pc33.example.com>")
	req := &Request{
		Method: MethodInvite,
		URI:    "sip:bob@example.com",
		To:     Address{Name: "Bob", URI: URI{Scheme: "sip", Opaque: "bob@example.com"}},
		From:   Address{Name: "Alice", URI: URI{Scheme: "sip", Opaque: "alice@example.com"}},
		Via:    Via{Address: "pc33.example.com", Branch: "776asdhds"},
		Header: header,
	}

	if _, err := WriteRequest(io.Discard, req); err != nil {
		t.Fatalf("write request: %v", err)
	}
}

func TestAddress(t *testing.T) {
	var tests = []struct {
		name string
		addr string
		want string
	}{
		{"bare", "sip:test@example.com", "<sip:test@example.com>"},
		{"basic", "<sip:test@example.com>", "<sip:test@example.com>"},
		{"bare tag", "sip:+1234@example.com;tag=887s", "<sip:+1234@example.com>;tag=887s"},
		{"tag", "<sip:test@example.com>;tag=1234", "<sip:test@example.com>;tag=1234"},
		{"name", "Oliver <sip:test@example.com>", "Oliver <sip:test@example.com>"},
		{"name tag", "Oliver <sip:test@example.com>;tag=1234", "Oliver <sip:test@example.com>;tag=1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddress(tt.addr)
			if err != nil {
				t.Fatalf("parse %q: %v", tt.addr, err)
			}
			if got.String() != tt.want {
				t.Fatalf("ParseAddress(%q) = %s, want %s", tt.addr, got, tt.want)
			}
		})
	}
}

func TestReadRequest(t *testing.T) {
	f, err := os.Open("testdata/invite")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = ReadRequest(f)
	if err != nil {
		t.Fatal("read request:", err)
	}

}

func TestResponse(t *testing.T) {
	raw := `SIP/2.0 200 OK
Via: SIP/2.0/UDP server10.example.com
   ;branch=z9hG4bKnashds8;received=192.0.2.3
Via: SIP/2.0/UDP bigbox3.site3.example.com
   ;branch=z9hG4bK77ef4c2312983.1;received=192.0.2.2
Via: SIP/2.0/UDP pc33.example.com
   ;branch=z9hG4bK776asdhds ;received=192.0.2.1
To: Bob <sip:bob@example.com>;tag=a6c85cf
From: Alice <sip:alice@example.com>;tag=1928301774
Call-ID: a84b4c76e66710@pc33.example.com
CSeq: 314159 INVITE
Contact: <sip:bob@192.0.2.4>
Content-Type: application/sdp
Content-Length: 131

...`
	msg, err := readMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatal("read message:", err)
	}
	_, err = parseResponse(msg)
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}
}
