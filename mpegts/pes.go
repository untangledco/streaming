package mpegts

import (
	"encoding/binary"
	"fmt"
)

// PESPacket represents a packetised elementary stream (PES) packet.
// These are transported in MPEG-TS packet payloads.
type PESPacket struct {
	// ID uniquely identifies this stream from other elementary streams.
	ID byte
	// Length is the number of bytes for this packet.
	// A value of zero indicates the packet can be of any length,
	// but is only valid when the underlying stream is video.
	Length uint16
	Header *PESHeader
	// Data contains raw audio/video data.
	Data []byte
}

// PESHeader represents the optional header of a PES packet.
type PESHeader struct {
	// Scrambling informs decoders how packet contents are encrypted or "scrambled".
	// Zero indicates the stream is not scrambled.
	Scrambling uint8
	// Priority signals to decoders that this packet should be processed before others.
	Priority bool
	// Alignment indicates thatheader is immediately followed by
	// the video start code or audio syncword.
	Alignment bool
	// Copyrighted signals that stream's content is copyrighted.
	Copyrighted bool
	// Original signals whether the stream is an original or copy.
	Original bool
	// Fields indicates which optional fields are present in Optional.
	Fields       uint8
	Presentation *Timestamp
	Decode       *Timestamp
	// Optional holds any bytes that we don't know how to decode (yet).
	// TOOD(otl): also contains stuffing bytes, which should be stored separately.
	Optional []byte
}

func (h PESHeader) packedLength() int {
	n := headerLength
	if h.Presentation != nil {
		n += 5 // packed timestamps are [5]byte
	}
	if h.Decode != nil {
		n += 5 // packed timestamps are [5]byte
	}
	return n + len(h.Optional)
}

// flags + fields + field length
const headerLength int = 3

// Flags indicating which optional fields are present in a PES header's payload.
const (
	FieldPTS uint8 = 1 << (7 - iota)
	FieldDTS
	FieldESCR
	FieldESRate
	FieldTrickMode
	FieldCopyInfo
	FieldCRC
	FieldExtension
)

var pesHeaderPrefix [3]byte = [3]byte{0, 0, 1}

func isPESPayload(payload []byte) bool {
	if len(payload) < 6 {
		return false // too short
	}
	var prefix [3]byte
	copy(prefix[:], payload[:3])
	if prefix == pesHeaderPrefix {
		return true
	}
	return false
}

func decodePES(buf []byte) (*PESPacket, error) {
	if !isPESPayload(buf) {
		return nil, fmt.Errorf("no PES packet")
	}
	var pes PESPacket
	pes.ID = buf[3]
	pes.Length = binary.BigEndian.Uint16(buf[4:6])
	buf = buf[6:]
	// is there a header to decode?
	if pes.Length >= 3 {
		header, err := decodePESHeader(buf)
		if err != nil {
			return nil, fmt.Errorf("decode header: %w", err)
		}
		pes.Header = header
		// how many bytes we read from buf
		read := pes.Header.packedLength()
		buf = buf[read:]
	}
	pes.Data = buf
	return &pes, nil
}

func encodePESPacket(p *PESPacket) ([]byte, error) {
	// length is startcode + id + length
	buf := make([]byte, 3+1+2)
	copy(buf[:3], pesHeaderPrefix[:])
	buf[3] = p.ID
	binary.BigEndian.PutUint16(buf[4:6], p.Length)
	if p.Header != nil {
		b, err := encodePESHeader(p.Header)
		if err != nil {
			return nil, fmt.Errorf("encode PES header: %w", err)
		}
		buf = append(buf, b...)
	}
	if p.Data != nil {
		buf = append(buf, p.Data...)
	}
	return buf, nil
}

func decodePESHeader(buf []byte) (*PESHeader, error) {
	if len(buf) < 3 {
		return nil, fmt.Errorf("short buffer length %d: need at least 3", len(buf))
	}
	if buf[0]&0xc0 == 0 {
		return nil, fmt.Errorf("decode header: bad marker bits")
	}
	var h PESHeader
	flags := buf[0]
	h.Scrambling = flags & 0b00110000
	h.Priority = flags&0b00001000 > 0
	h.Alignment = flags&0b00000100 > 0
	h.Copyrighted = flags&0b00000010 > 0
	h.Original = flags&0b00000001 > 0

	h.Fields = buf[1]

	hlength := int(buf[2])
	if len(buf[3:]) < hlength {
		return nil, fmt.Errorf("short buffer: header reports %d, have %d", hlength, len(buf[3:]))
	}
	buf = buf[3 : 3+hlength]

	if h.Fields&FieldPTS > 0 {
		var tstamp Timestamp
		tstamp.PTS = h.Fields&FieldPTS > 0
		tstamp.DTS = h.Fields&FieldDTS > 0
		if !tstamp.PTS {
			return nil, fmt.Errorf("timestamp present but missing PTS")
		}
		var err error
		var a [5]byte
		copy(a[:], buf[:5])
		tstamp, err = unpackTimestamp(a)
		if err != nil {
			return nil, fmt.Errorf("read timestamp: %w", err)
		}
		h.Presentation = &tstamp
		buf = buf[5:]
	}
	if h.Fields&FieldDTS > 0 {
		var tstamp Timestamp
		tstamp.PTS = h.Fields&FieldPTS > 0
		tstamp.DTS = h.Fields&FieldDTS > 0
		if !tstamp.PTS {
			return nil, fmt.Errorf("timestamp present but missing PTS")
		}
		var err error
		var a [5]byte
		copy(a[:], buf[:5])
		tstamp, err = unpackTimestamp(a)
		if err != nil {
			return nil, fmt.Errorf("read timestamp: %w", err)
		}
		h.Decode = &tstamp
		buf = buf[5:]
	}
	h.Optional = buf
	return &h, nil
}

func encodePESHeader(h *PESHeader) ([]byte, error) {
	// length of 3: flags + fields + header length
	buf := make([]byte, 3)
	buf[0] |= (1 << 7) // marker bits
	buf[0] |= h.Scrambling
	if h.Priority {
		buf[0] |= (1 << 3)
	}
	if h.Alignment {
		buf[0] |= (1 << 2)
	}
	if h.Copyrighted {
		buf[0] |= (1 << 1)
	}
	if h.Original {
		buf[0] |= 1
	}
	buf[1] = h.Fields
	var opt []byte
	if h.Presentation != nil {
		if !h.Presentation.PTS && h.Presentation.DTS {
			return nil, fmt.Errorf("bad timestamp: DTS set without PTS")
		}
		packed := packTimestamp(*h.Presentation)
		opt = append(opt, packed[:]...)
	}
	if h.Decode != nil {
		if !h.Decode.PTS && h.Decode.DTS {
			return nil, fmt.Errorf("bad timestamp: DTS set without PTS")
		}
		packed := packTimestamp(*h.Decode)
		opt = append(opt, packed[:]...)
	}
	if h.Optional != nil {
		opt = append(opt, h.Optional...)
	}
	// TODO(otl): decode, encode stuffing
	buf[2] = byte(len(opt))
	buf = append(buf, opt...)
	return buf, nil
}

// Timestamp represents the timestamp transported in a PES packet
// header. It is used by decoders to reliably time and synchronise
// playback of video/audio payloads.
type Timestamp struct {
	// PTS indicates the packet carrying the timestamp contains a presentation timestamp.
	PTS bool
	// DTS indicates the packet carrying this timestamp contains a decode timestamp.
	DTS bool
	// Ticks holds a 33-bit integer counting the number of ticks of a 90KHz clock.
	Ticks uint64
}

const maxTicks uint64 = 0x1ffffffff // max 33-bit integer

// unpackTimestamp unpacks the Timestamp stored in a.
// It is packed in 40 bits across 5 bytes, shown in the following bitfield diagram:
//
//	0 00pd ttt1
//	1 tttt tttt
//	2 tttt ttt1
//	3 tttt tttt
//	4 tttt ttt1
//
// 0, 1, 2, 3, 4 are each byte index in the array a.
// The two leading 0s are padding bits.
// P and D hold the values for PTS and DTS in Timestamp.
// The 1s are check bits which must be toggled.
// Ts are the 33-bit big-endian encoded integer.
func unpackTimestamp(a [5]byte) (Timestamp, error) {
	var tstamp Timestamp
	tstamp.PTS = a[0]&0b00100000 > 0
	tstamp.DTS = a[0]&0b00010000 > 0
	if tstamp.DTS && !tstamp.PTS {
		return Timestamp{}, fmt.Errorf("DTS set but no PTS set")
	}
	if a[0]&a[2]&a[4]&0x01 == 0 {
		return Timestamp{}, fmt.Errorf("corrupt timestamp")
	}

	tbuf := make([]byte, 5) // enough for 33-bit integer
	tbuf[0] = (a[0] & 0b00001000) >> 3

	tbuf[1] = (a[0] & 0b00000110) << 5
	tbuf[1] |= (a[1] & 0b11111100) >> 2

	tbuf[2] = (a[1] & 0b00000011) << 6
	tbuf[2] |= (a[2] & 0b11111100) >> 2

	tbuf[3] = (a[2] & 0b00000010) << 6
	tbuf[3] |= (a[3] & 0b11111110) >> 1

	tbuf[4] = (a[3] & 0b00000001) << 7
	tbuf[4] |= (a[4] & 0b11111110) >> 1
	buf := make([]byte, 8)
	copy(buf[3:], tbuf)
	tstamp.Ticks = binary.BigEndian.Uint64(buf)
	return tstamp, nil
}

// packTimestamp returns an array containing a packed Timestamp.
// The Timestamp is packed according to the bitfield layout documented
// in unpackTimestamp.
func packTimestamp(t Timestamp) [5]byte {
	var a [5]byte
	if t.PTS {
		a[0] |= (1 << 5)
	}
	if t.DTS {
		a[0] |= (1 << 4)
	}
	ticks := make([]byte, 8) // sizeof uint64
	binary.BigEndian.PutUint64(ticks, t.Ticks)
	// we don't want the whole 64-bit integer, only enough for 33 bits.
	ticks = ticks[3:]

	b := ticks[0] & 0b00000001
	a[0] |= (b << 3)

	// ticks[0] is packed.
	// 2 bits remaining before check bit in a[0].
	b = ticks[1] & 0b11000000
	a[0] |= (b >> 5)
	a[0] |= 1 // toggle check bit

	// 6 bits remaining in ticks[1].
	b = ticks[1] & 0b00111111
	a[1] |= (b << 2)

	// 2 bits remaining in a[1].
	// we're done with ticks[1], so pack left-most 2 bits from ticks[2].
	b = ticks[2] & 0b11000000
	a[1] |= (b >> 6)

	// a[1] packed.
	// pack remaining 6 bits from ticks[2] into a[2].
	b = ticks[2] & 0b00111111
	a[2] |= (b << 2)

	// ticks[2] done.
	// 1 bit remaining in a[2] before the check bit.
	// pack 1 bit from ticks[3] into a[2].
	b = ticks[3] & 0b10000000
	a[2] |= (b >> 6)
	a[2] |= 1 // toggle check bit

	// 7 bits remaining in ticks[3].
	// a[2] done.
	b = ticks[3] & 0b01111111
	a[3] |= (b << 1)

	// ticks[3] all packed.
	// a[3] has 1 bit remaining.
	b = ticks[4] & 0b10000000
	a[3] |= (b >> 7)

	// a[3] done.
	// 7 bits remaining to pack from ticks[4].
	b = ticks[4] & 0b01111111
	a[4] = (b << 1)

	// just 1 bit left to pack from ticks[4]! woo hoo!
	// just 1 bit free in a[4] before the check bit.
	b = ticks[4] & 0b10000000
	a[4] |= (b >> 7)
	a[4] |= 1 // toggle check bit
	return a
}
