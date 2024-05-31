package scte35_test

import (
	"encoding/base64"
	"fmt"

	"github.com/untangledco/streaming/scte35"
)

func Example() {
	msg := "/DA0AAAAAAAA///wBQb+cr0AUAAeAhxDVUVJSAAAjn/PAAGlmbAICAAAAAAsoKGKNAIAmsnRfg=="
	b, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		// handle error...
	}
	splice, err := scte35.Decode(b)
	if err != nil {
		// handle error...
	}

	dur := splice.Descriptors[0].(scte35.SegmentationDescriptor).Duration
	fmt.Println(*dur)
	*dur += 15 * 90000 // add 15 seconds as per 90KHz clock.
	fmt.Println(*splice.Descriptors[0].(scte35.SegmentationDescriptor).Duration)

	bb, err := scte35.Encode(splice)
	if err != nil {
		// handle error...
	}
	fmt.Println(base64.StdEncoding.EncodeToString(bb))

	// Output: 27630000
	// 28980000
	// /DA0AAAAAAAA///wBQb+cr0AUAAeAhxDVUVJSAAAjn/PAAG6MyAICAAAAAAsoKGKNAIAQO/SuQ==
}

func ExampleEncode() {
	when := uint64(12 * 60 * 60 * 90000) // 12 hours since midnight UTC as 90KHz ticks
	duration := uint64(60 * 90000)       // 60 seconds as 90KHz ticks
	splice := scte35.Splice{
		SAPType: scte35.SAPNone,
		Tier:    0x0fff,
		Command: &scte35.Command{
			Type:       scte35.TimeSignal,
			TimeSignal: &when,
		},
		Descriptors: []scte35.SpliceDescriptor{
			scte35.SegmentationDescriptor{
				EventID:      1234,
				Restrictions: scte35.NoRegionalBlackout | scte35.ArchiveAllowed | scte35.DeviceRestrictionsNone,
				Duration:     &duration,
				UPID: scte35.UPID{
					Type:  scte35.UPIDTI,
					Value: []byte{0x00, 0x00, 0x00, 0x00, 0x2c, 0xa0, 0xa1, 0x8a},
				},
				Type:     0x34,
				Number:   2,
				Expected: 0,
			},
		},
	}
	b, err := scte35.Encode(&splice)
	if err != nil {
		// handle error...
	}
	// SCTE 35 time_signal() commands can be inserted into HLS playlists using the
	// EXT-X-DATERANGE tag. See "Mapping SCTE-35 into EXT-X-DATERANGE"
	// RFC 8216 section 4.3.2.7.1.
	fmt.Printf("#EXT-X-DATERANGE:ID=\"example\",SCTE35-CMD=%#x\n", b)
}
