package sdp_test

import (
	"fmt"
	"net/netip"

	"github.com/untangledco/streaming/sdp"
)

// The simplest Session may be described by setting the mandatory fields Origin and Name.
// The Session type implements fmt.Stringer.
// To encode a Session in the SDP text format, use Session.String().
func Example() {
	session := sdp.Session{
		Origin: sdp.Origin{
			ID:      3930287268, // example only; use sdp.Now()
			Version: 3930287268, // example only; use sdp.Now()
			Address: netip.MustParseAddr("2001:db8::1"),
		},
		Name: "A call from me to you",
	}
	fmt.Printf("%s", session)
	// Output:
	// v=0
	// o=- 3930287268 3930287268 IN IP6 2001:db8::1
	// s=A call from me to you
	// t=0 0
}
