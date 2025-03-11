package sip

import (
	"strings"
	"testing"
)

func TestReadRequest(t *testing.T) {
	raw := `INVITE sip:bob@biloxi.com SIP/2.0
Via: SIP/2.0/UDP pc33.atlanta.com;branch=z9hG4bK776asdhds
Max-Forwards: 70
To: Bob <sip:bob@biloxi.com>
From: Alice <sip:alice@atlanta.com>;tag=1928301774
Call-ID: a84b4c76e66710@pc33.atlanta.com
CSeq: 314159 INVITE
Contact: <sip:alice@pc33.atlanta.com>
Content-Type: application/sdp
Content-Length: 328

...`

	_, err := ReadRequest(strings.NewReader(raw))
	if err != nil {
		t.Fatal("read request:", err)
	}
}
