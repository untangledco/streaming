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
