package mpegts

import (
	"encoding/binary"
	"fmt"
	"io"
)

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
	if buf[0] != Sync {
		return nil, fmt.Errorf("expected sync byte, got %x", buf[0])
	}
	var p Packet
	if buf[1]&0x80 > 0 {
		p.Error = true
	}
	if buf[1]&0x40 > 0 {
		p.PayloadStart = true
	}
	if buf[1]&0x20 > 0 {
		p.Priority = true
	}
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
		p.Payload = buf[4:]
	case 0x03:
		p.Adaptation = buf[4:]
	default:
		return &p, fmt.Errorf("cannot decode both payload and adaptation field")
	}
	return &p, nil
}
