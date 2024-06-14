package mpegts_test

import (
	"fmt"
	"log"
	"os"

	"github.com/untangledco/streaming/mpegts"
)

// The most common usage of mpegts involves the use of Scanner which
// steps through the packets of a transport stream. In this example we
// decode then re-encode every packet from the standard input to the
// standard output.
// If a packet contains a clock reference, we print that out to the standard
// error for diagnostics.
// This provides similar functionality to the following ffprobe command:
//
//	ffprobe -show_packets -
func Example() {
	sc := mpegts.NewScanner(os.Stdin)
	var i int
	for sc.Scan() {
		i++
		packet := sc.Packet()
		if packet.Adaptation != nil && packet.Adaptation.PCR != nil {
			ticks := packet.Adaptation.PCR.Ticks()
			fmt.Fprintf(os.Stderr, "packet %d\t%d\n", i, ticks)
		}
		if err := mpegts.Encode(os.Stdout, packet); err != nil {
			log.Printf("encode packet %d: %v", i, err)
		}
	}
	if sc.Err() != nil {
		log.Fatalf("scan: %v", sc.Err())
	}
}
