package sdp_test

import (
	"fmt"
	"net/netip"

	"github.com/untangledco/streaming/rtp"
	"github.com/untangledco/streaming/sdp"
)

// A simple Session describing transmission of uncompressed linear PCM
// audio in RTP starts by setting the mandatory fields Origin and Name.
// The Media field contains information about the audio such as the sample rate
// and the number of audio channels.
// The Session type implements fmt.Stringer;
// to encode a Session in the SDP text format, use Session.String().
func Example() {
	session := sdp.Session{
		Origin: sdp.Origin{
			ID:      3930287268, // example only; use sdp.Now()
			Version: 3930287268, // example only; use sdp.Now()
			Address: netip.MustParseAddr("2001:db8::1"),
		},
		Name: "A call from me to you",
		Media: []sdp.Media{
			{
				Type:      sdp.MediaTypeAudio,
				Port:      6969,
				Transport: sdp.ProtoRTP,
				Format:    []string{fmt.Sprintf("%d", rtp.PayloadL16Mono)},
				Attributes: []string{
					fmt.Sprintf("rtpmap:%d", rtp.PayloadL16Mono),
					fmt.Sprintf("L16/%d", 22050),
				},
			},
		},
	}
	fmt.Printf("%s", session)
	// Output:
	// v=0
	// o=- 3930287268 3930287268 IN IP6 2001:db8::1
	// s=A call from me to you
	// t=0 0
	// m=audio 6969 RTP/AVP 11
	// a=rtpmap:11 L16/22050
}
