package pcap

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestDecode(t *testing.T) {
	f, err := os.Open("testdata/text_udp.pcap")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gheader := GlobalHeader{
		SnapLen: 524288,
	}
	header := Header{
		// from tcpdump -tt -r testdata/text_udp.pcap
		Time:    time.UnixMicro(1721314372204926),
		OrigLen: 45,
	}

	capture, err := decode(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(capture.Packets) != 1 {
		t.Fatalf("expected 1 packet, found %d", len(capture.Packets))
	}
	if capture.Header != gheader {
		t.Errorf("decoded global header %v, want %v", capture.Header, gheader)
	}
	if capture.Packets[0].Header != header {
		t.Errorf("decoded packet header %v, want %v", capture.Packets[0].Header, header)
	}
}

func TestEncode(t *testing.T) {
	want, err := os.ReadFile("testdata/text_udp.pcap")
	if err != nil {
		t.Fatal(err)
	}
	capture, err := decode(bytes.NewReader(want))
	if err != nil {
		t.Fatal(err)
	}
	got, err := encode(capture)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("encode(%v) = %x, want %x", capture, got, want)
	}
}

func TestTimestamp(t *testing.T) {
	want := [2]uint32{1, 100}
	when := time.Unix(1, 100)
	sec, nsec := timestamp(when)
	got := [2]uint32{sec, nsec}
	if got != want {
		t.Errorf("timestamp(%s) = %v, want %v", when, got, want)
	}
}
