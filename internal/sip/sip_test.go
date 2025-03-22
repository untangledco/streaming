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
	header.Set("Call-ID", "blabla")
	header.Set("To", "test <sip:test@example.com>")
	header.Set("From", "Oliver <sip:o@olowe.co>")
	header.Set("CSeq", "1 "+MethodRegister)
	req := &Request{
		Method: MethodRegister,
		URI:    "sip:test@example.com",
		Header: header,
	}
	_, err := WriteRequest(io.Discard, req)
	if err == nil {
		t.Errorf("no error writing request with zero Via field")
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
