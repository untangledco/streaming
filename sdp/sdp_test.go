package sdp

import (
	"net/mail"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestReadSession(t *testing.T) {
	s := `
v=0
o=jdoe 3724394400 3724394405 IN IP4 198.51.100.1
s=Call to John Smith
i=SDP Offer #1
u=http://www.jdoe.example.com/home.html
e=Jane Doe <jane@jdoe.example.com>
p=+1 617 555-6011
c=IN IP4 198.51.100.1
t=0 0
m=audio 49170 RTP/AVP 0
m=audio 49180 RTP/AVP 0
m=video 51372 RTP/AVP 99
c=IN IP6 2001:db8::2
a=rtpmap:99 h263-1998/90000`
	want := Session{
		Name: "Call to John Smith",
		Origin: Origin{
			Username:    "jdoe",
			ID:          3724394400,
			Version:     3724394405,
			Network:     "IN",
			AddressType: "IP4",
			Address:     "198.51.100.1",
		},
		Info: "SDP Offer #1",
		URI: &url.URL{Scheme: "http",
			Host: "www.jdoe.example.com",
			Path: "/home.html",
		},
		Email: &mail.Address{"Jane Doe", "jane@jdoe.example.com"},
		Phone: "+1 617 555-6011",
	}
	session, err := ReadSession(strings.NewReader(s))
	if err != nil {
		t.Fatalf("read session: %v", err)
	}
	if !reflect.DeepEqual(*session, want) {
		t.Errorf("got %+v\nwant %+v\n", *session, want)
	}
	t.Errorf("TODO still not parsing all fields")
}
