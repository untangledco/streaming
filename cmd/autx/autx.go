// Command autx transmits audio over the network via RTP.
// It reads raw audio from the standard input
// and transmits it to the provided address.
// Audio must be single-channel (mono),
// signed 16-bit linear PCM data in big-endian byte order.
// The sample rate must be 44.1KHz.
//
// Its usage is:
//
//	autx address
//
// Address is in addr:port format.
//
// # Example
//
// Transmit audio from the file "test.pcm" to 2001:db8::1 port 9999:
//
//	autx [2001:db8::1]:9999 < test.pcm
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/untangledco/streaming/rtp"
	"github.com/untangledco/streaming/sdp"
)

const sampleRate = 4000
const bitsPerSample = 8
const tick = 500 // milliseconds
const packetInterval = tick * time.Millisecond
const bufSize = sampleRate * bitsPerSample / 8 / 1000 * tick

const usage string = "usage: autx address"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	session, err := rtp.Dial("udp", os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	session.Clock = 44100 // 44.1KHz, not the audio sample rate.

	description := sdp.Session{
		Origin: sdp.Origin{
			Username:    sdp.NoUsername,
			ID:          sdp.Now(),
			Version:     sdp.Now(),
			AddressType: "IP6",
			Address:     "::1",
		},
		Name: "test",
		Media: []sdp.Media{
			{
				Type:      sdp.MediaTypeAudio,
				Port:      9999,
				Transport: sdp.ProtoRTP,
				Format:    []string{fmt.Sprintf("%d", rtp.PayloadType(99))},
				Attributes: []string{
					fmt.Sprintf("rtpmap:%d", rtp.PayloadType(99)),
					fmt.Sprintf("L8/%d", sampleRate),
				},
			},
		},
	}
	fmt.Fprintln(os.Stderr, description)

	out := make(chan rtp.Packet, 100)
	go func() {
		buf := &bytes.Buffer{}
		p := rtp.Packet{
			Header: rtp.Header{
				Type: rtp.PayloadL16,
			},
		}
		for {
			buf.Reset()
			_, err := io.CopyN(buf, os.Stdin, int64(bufSize))
			if errors.Is(err, io.EOF) {
				p.Payload = buf.Bytes()
				out <- p
				close(out)
				return
			} else if err != nil {
				log.Println(err)
			}
			p.Payload = buf.Bytes()
			out <- p
		}
	}()

	for p := range out {
		if err := session.Transmit(&p); err != nil {
			log.Println(err)
		}
	}
}
