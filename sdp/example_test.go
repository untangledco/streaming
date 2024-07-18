package sdp_test

import (
	"fmt"

	"github.com/untangledco/streaming/sdp"
)

// The simplest Session may be described by setting the mandatory fields Origin and Name.
// The Session type implements fmt.Stringer.
// To encode a Session in the SDP text format, use Session.String().
func Example() {
	session := sdp.Session{
		Origin: sdp.Origin{
			Username:    sdp.NoUsername,
			ID:          3930287268, // example only; use sdp.Now()
			Version:     3930287268, // example only; use sdp.Now()
			AddressType: "IP6",      // or "IP4"
			Address:     "2001:db8::1",
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
