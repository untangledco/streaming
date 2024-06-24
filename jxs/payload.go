// Package jxs ... for transporting JPEG XS streams in RTP payloads
// as specified in RFC 9134.
package jxs

import "encoding/binary"

type ScanMode uint8

const (
	ScanModeProgressive ScanMode = 0
	_
	ScanModeInterlacedFirst  ScanMode = 0b00010000
	ScanModeInterlacedSecond          = 0b00011000
)

/*
type Payload struct {
	Header            *Header
	PacketizationUnit *PacketizationUnit
}

type PacketizationUnit struct {
	VideoBox
	ColorBox
	// ...
}
*/

type Header struct {
	// Sequential indicates whether packets are sent sequentially
	// or possibly out of order. may be sent out of order.
	Sequential bool

	// PacketMode indicates the packetization mode. If true,
	// codestream packetization is used. If false, slice
	// packetization is used and Sequential must be false.
	PacketMode bool

	// Last indicates whether the packet is the last packet of a
	// packetization unit.
	Last bool

	// InterlacedInfo specifies how the frame is scanned.
	InterlacedInfo ScanMode

	// FrameCount holds a 5-bit integer (modulo 32) identifying
	// the frame to which the packet belongs.
	FrameCount uint8

	// SEPCounter (Slice and Extended Packet counter)
	// holds an 11-bit integer which is interpreted differently
	// depending on the value of PacketMode.
	//
	// If the packetization mode is codestream, then SEPCounter is
	// incremented by 1 whenever PacketCounter overruns.
	//
	// If packetization mode is slice, then SEPCounter identifies
	// the slice (modulo 2047) to which the packet contributes.
	SEPCount uint16

	// PacketCounter is an 11-bit integer identifying the packet number
	// within the current packetization unit.
	PacketCount uint16
}

func unmarshalHeader(a [4]byte, hdr Header) error {
	hdr.Sequential = a[0]&0b10000000 > 0
	hdr.PacketMode = a[0]&0b01000000 > 0
	hdr.Last = a[0]&0b00100000 > 0
	hdr.InterlacedInfo = ScanMode(a[0] & 0b00011000)

	hdr.FrameCount = (a[0] & 0b00000111) << 2
	hdr.FrameCount |= ((a[1] & 0b11000000) >> 6)

	var sepc [2]byte
	sepc[0] = a[1] & 0b00111000 >> 3
	sepc[1] = a[1] & 0b00000111 << 5
	sepc[1] |= a[2] & 0b11111000 >> 3
	hdr.SEPCount = binary.BigEndian.Uint16(sepc[:])

	hdr.PacketCount = binary.BigEndian.Uint16([]byte{a[2] & 0b00000111, a[3]})
	return nil
}

func marshalHeader(hdr Header) [4]byte {
	var a [4]byte
	if hdr.Sequential {
		a[0] |= (1 << 7)
	}
	if hdr.PacketMode {
		a[0] |= (1 << 6)
	}
	if hdr.Last {
		a[0] |= (1 << 5)
	}
	a[0] |= byte(hdr.InterlacedInfo)

	// Need to pack the 5 bits in FrameCount across both a[0] and a[1].
	// Have 3 bits remaining in a[0].
	a[0] |= (hdr.FrameCount & 0b00011100) >> 2
	// Pack remaining 2 bits from FrameCount into left-most bits of a[1].
	a[1] = (hdr.FrameCount & 0b00000011) << 6

	// 6 bits remaining in a[1].
	// We have 11 bits in SEPCount to pack.
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, hdr.SEPCount)
	a[1] |= (b[0] & 0b00000111) << 3
	// 3 bits remaining in a[1], so get next 3 bits from b.
	a[1] |= (b[1] & 0b11100000) >> 5

	// 5 bits remaining in b.
	a[2] = (b[1] & 0b00011111) << 3

	b[0] = 0
	b[1] = 0
	binary.BigEndian.PutUint16(b, hdr.PacketCount)
	// 3 bits spare in a[2].
	a[2] |= (b[0] & 0b00000111)
	a[3] = b[1]
	return a
}
