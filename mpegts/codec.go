package mpegts

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var ErrLongPacket = errors.New("long packet")
var ErrShortPacket = errors.New("short packet")

func Unmarshal(buf []byte, p *Packet) error {
	if len(buf) != PacketSize {
		return fmt.Errorf("need exactly %d bytes, have %d", PacketSize, len(buf))
	}
	if buf[0] != Sync {
		return fmt.Errorf("expected sync byte, got %x", buf[0])
	}
	p.Error = (buf[1] & 0x80) > 0
	p.PayloadStart = (buf[1] & 0x40) > 0
	p.Priority = (buf[1] & 0x20) > 0
	// Want next 13 bits. 5 from buf[1] and all of buf[2].
	pid := binary.BigEndian.Uint16([]byte{buf[1] & 0x1f, buf[2]})
	p.PID = PacketID(pid)

	// next 2 bits
	p.Scrambling = Scramble(buf[3] & 0xc0)
	// skip next 2 bits until later when we need to decode adaptation or payload.
	// now just get last 4 bits
	p.Continuity = buf[3] & 0x0f

	afc := buf[3] >> 4
	switch afc {
	case 0x01:
		buf = buf[4:]
	case 0x02, 0x03:
		p.Adaptation = parseAdaptationField(buf[4:])
		if p.Adaptation == nil {
			p.emptyAdaptation = true
		}
		alen := int(buf[4])
		buf = buf[4+1+alen:]
	default:
		return fmt.Errorf("neither adaptation field or payload present")
	}

	return unmarshalPayload(buf, p)
}

func Decode(r io.Reader) (*Packet, error) {
	buf := make([]byte, PacketSize)
	n, err := r.Read(buf)
	if n != PacketSize {
		if err != nil {
			return nil, fmt.Errorf("short read (%d bytes): %w", n, err)
		}
		return nil, fmt.Errorf("short read (%d bytes)", n)
	}
	if err != nil {
		return nil, err
	}
	var p Packet
	if err := Unmarshal(buf, &p); err != nil {
		return &p, fmt.Errorf("unmarshal packet: %w", err)
	}
	return &p, nil
}

func parseAdaptationField(buf []byte) *Adaptation {
	length := int(buf[0])
	if length == 0 {
		return nil
	}
	buf = buf[1 : length+1]
	var af Adaptation
	flags := buf[0]
	buf = buf[1:]
	af.Discontinuous = flags&0x80 > 0
	af.RandomAccess = flags&0x40 > 0
	af.Priority = flags&0x20 > 0
	if flags&0x10 > 0 {
		var p [6]byte
		copy(p[:], buf[:6])
		pcr := parsePCR(p)
		af.PCR = &pcr
		buf = buf[6:]
	}
	if flags&0x08 > 0 {
		var p [6]byte
		copy(p[:], buf[:6])
		pcr := parsePCR(p)
		af.OPCR = &pcr
		buf = buf[6:]
	}
	if flags&0x04 > 0 {
		af.SpliceCountdownSet = true
		af.SpliceCountdown = buf[0]
		buf = buf[1:]
	}
	if flags&0x02 > 0 {
		tlen := int(buf[0])
		af.Private = buf[1:tlen]
		buf = buf[tlen:]
	}
	if flags&0x01 > 0 {
		extlen := int(buf[0])
		af.Extension = buf[1:extlen]
		buf = buf[extlen:]
	}
	if len(buf) > 0 {
		af.Stuffing = buf
	}
	return &af
}

// parsePCR parses the encoded PCR from a.
// The 33-bit base and the 9-bit extension
// are stored in a 6 byte array with the following bit layout,
// where "b" stands for "base", "r" for reserved bits, and "e" for extension.
//
//	0 bbbb bbbb
//	1 bbbb bbbb
//	2 bbbb bbbb
//	3 bbbb bbbb
//	4 brrr rrre
//	5 eeee eeee
func parsePCR(a [6]byte) PCR {
	// we only want the left-most bit.
	// 6 bits are reserved and the right-most bit is part of extension.
	b := [8]byte{0, 0, 0, a[0], a[1], a[2], a[3], a[4] & 0x80}
	base := binary.BigEndian.Uint64(b[:])
	base = base >> 7 // trim masked reserved, extension bits
	// next 6 bits of a[4] are reserved, so right-most bit in a[4]
	// and all of a[5] have the extension.

	ext := binary.BigEndian.Uint16([]byte{a[4] & 0x01, a[5]})
	return PCR{base, ext}
}

func unmarshalPayload(payload []byte, p *Packet) error {
	if isPESPayload(payload) && p.PayloadStart {
		pes, err := decodePES(payload)
		if err != nil {
			return fmt.Errorf("decode PES packet: %w", err)
		}
		p.PES = pes
	} else {
		p.Payload = payload
	}
	return nil
}

func Encode(w io.Writer, p *Packet) error {
	buf := make([]byte, 4)
	buf[0] = Sync
	if p.Error {
		buf[1] |= 0x80
	}
	if p.PayloadStart {
		buf[1] |= 0x40
	}
	if p.Priority {
		buf[1] |= 0x20
	}
	if p.PID > PacketNull {
		return fmt.Errorf("packet id %s greater than max %s", p.PID, PacketNull)
	}
	buf[1] |= byte(p.PID >> 8)
	buf[2] = byte(p.PID)

	buf[3] |= byte(p.Scrambling)
	if p.Adaptation != nil || p.emptyAdaptation {
		buf[3] |= 0x20
	}
	if p.Payload != nil || p.PES != nil {
		buf[3] |= 0x10
	}
	if p.Continuity > 15 {
		return fmt.Errorf("continuity %d larger than max 4-bit integer %d", p.Continuity, 15)
	}
	buf[3] |= p.Continuity

	if p.Adaptation != nil {
		alen := 1 // just flags
		if p.Adaptation.PCR != nil {
			alen += 6
		}
		if p.Adaptation.OPCR != nil {
			alen += 6
		}
		if p.Adaptation.SpliceCountdownSet {
			alen++ // single byte
		}
		if p.Adaptation.Private != nil {
			alen++ // 1 byte to store length of private
			alen += len(p.Adaptation.Private)
		}
		alen += len(p.Adaptation.Extension)
		alen += len(p.Adaptation.Stuffing)
		if alen > 255 {
			return fmt.Errorf("adaptation field too long: have %d bytes, max %d", alen, 255)
		}

		abuf := make([]byte, 1+alen) // length + total
		abuf[0] = uint8(alen)
		var i int = 2 // cursor; after length and flags
		if p.Adaptation.Discontinuous {
			abuf[1] |= 0x80
		}
		if p.Adaptation.RandomAccess {
			abuf[1] |= 0x40
		}
		if p.Adaptation.Priority {
			abuf[1] |= 0x20
		}
		if p.Adaptation.PCR != nil {
			abuf[1] |= 0x10
			if err := putPCR(abuf[i:i+6], p.Adaptation.PCR); err != nil {
				return fmt.Errorf("pack PCR: %w", err)
			}
			i += 6
		}
		if p.Adaptation.OPCR != nil {
			abuf[1] |= 0x08
			if err := putPCR(abuf[i:i+6], p.Adaptation.OPCR); err != nil {
				return fmt.Errorf("pack OPCR: %w", err)
			}
			i += 6
		}
		if p.Adaptation.SpliceCountdownSet {
			abuf[1] |= 0x04
			abuf[i] = p.Adaptation.SpliceCountdown
			i++
		}
		if p.Adaptation.Private != nil {
			abuf[1] |= 0x02
			if len(p.Adaptation.Private) > 255 {
				return fmt.Errorf("private data length %d longer than max %d", len(p.Adaptation.Private), 255)
			}
			abuf[i] = byte(len(p.Adaptation.Private))
			i++
			copy(abuf[i:], p.Adaptation.Private)
			i += len(p.Adaptation.Private)
		}
		if p.Adaptation.Extension != nil {
			abuf[1] |= 0x01
			copy(abuf[i:], p.Adaptation.Extension)
			i += len(p.Adaptation.Extension)
		}
		if p.Adaptation.Stuffing != nil {
			copy(abuf[i:], p.Adaptation.Stuffing)
		}
		buf = append(buf, abuf...)
	} else if p.emptyAdaptation {
		// no adaptation field to encode, but we need to store an adaptation field length of 0.
		buf = append(buf, 0)
	}
	if p.PES != nil {
		b, err := encodePESPacket(p.PES)
		if err != nil {
			return fmt.Errorf("encode PES packet: %w", err)
		}
		buf = append(buf, b...)
	}
	if p.Payload != nil {
		buf = append(buf, p.Payload...)
	}
	if len(buf) > PacketSize {
		return fmt.Errorf("%w: %d bytes", ErrLongPacket, len(buf))
	} else if len(buf) < PacketSize {
		return fmt.Errorf("%w: %d bytes", ErrShortPacket, len(buf))
	}
	_, err := w.Write(buf)
	return err
}

const (
	baseMax      = 8589934592 - 1 // max 33-bit uint
	extensionMax = 512 - 1        // max 9-bit uint
)

func putPCR(b []byte, pcr *PCR) error {
	if len(b) != 6 {
		return fmt.Errorf("need %d bytes, got %d", 6, len(b))
	}
	if pcr.Base > baseMax {
		return fmt.Errorf("base %d larger than max %d", pcr.Base, baseMax)
	} else if pcr.Extension > extensionMax {
		return fmt.Errorf("extension %d larger than max %d", pcr.Extension, extensionMax)
	}

	ubuf := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(ubuf, pcr.Base)
	// we're only working with 33 bits, so slice off the first 3
	// bytes to get 4 + 1 bytes (32+1 bits)
	ubuf = ubuf[3:]

	// now pack 33 bits from ubuf into b[:4].
	// The 33rd bit of our 33-bit integer is in the first byte: 0b00000001.
	// We're packing bits from left to right, so shift it left and assign to b[0].
	b[0] = ubuf[0] << 7

	// We have 7 bits free in b[0], so get 7 bits from ubuf[1] and pack it into b[0].
	b[0] |= ubuf[1] >> 1

	// 1 bit left in ubuf[1]; put it in the next dest byte.
	// Rinse and repeat until we're out of bits.
	b[1] = ubuf[1] << 7
	b[1] |= ubuf[2] >> 1
	b[2] = ubuf[2] << 7
	b[2] |= ubuf[3] >> 1
	b[3] = ubuf[3] << 7
	b[3] |= ubuf[4] >> 1
	b[4] = ubuf[4] << 7
	// No more base bits to pack.

	// Next, toggle the 6 reserved bits.
	b[4] |= 0b01111110

	var ext [2]byte
	binary.BigEndian.PutUint16(ext[:], pcr.Extension)
	b[4] |= ext[0]
	b[5] = ext[1]
	return nil
}
