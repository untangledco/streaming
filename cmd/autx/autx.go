// Command autx transmits audio over the network via RTP.
// It reads raw audio from the standard input
// and transmits it to the provided address every 30 milliseconds.
// Audio must be single-channel (mono),
// signed 16-bit linear PCM data in big-endian byte order.
// The sample rate must be 22.05KHz.
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
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/untangledco/streaming/rtp"
	"github.com/untangledco/streaming/sdp"
)

const sampleRate = 22050
const bitsPerSample = 16
const tick = 30 // milliseconds
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

	origin := sdp.Origin{
		ID:      sdp.Now(),
		Version: sdp.Now(),
		Address: netip.AddrFrom4([4]byte{127, 0, 0, 1}),
	}
	if strings.HasPrefix(os.Args[1], "[") {
		origin.Address = netip.IPv6Loopback()
	}
	_, port, err := net.SplitHostPort(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	nport, err := strconv.Atoi(port)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse port:", err)
		os.Exit(1)
	}

	description := sdp.Session{
		Origin: origin,
		Name:   "test",
		Media: []sdp.Media{
			{
				Type:      sdp.MediaTypeAudio,
				Port:      nport,
				Transport: sdp.ProtoRTP,
				Format:    []string{fmt.Sprintf("%d", rtp.PayloadType(11))},
				Attributes: []string{
					fmt.Sprintf("rtpmap:%d", rtp.PayloadType(11)),
					fmt.Sprintf("L16/%d", sampleRate),
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
				Type: rtp.PayloadType(11),
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
			p.Payload = bytes.Clone(buf.Bytes())
			out <- p
		}
	}()

	for range time.NewTicker(packetInterval).C {
		p, ok := <-out
		if !ok {
			return
		}
		if err := session.Transmit(&p); err != nil {
			log.Println(err)
		}
	}
}
