package rtp

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func Dial(network, addr string) (*Session, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	s := Session{conn: conn}
	s.init()
	return &s, nil
}

// Session represents a RTP session... TODO(otl)
// When a Session is established with Dial(), ...
type Session struct {
	Type PayloadType
	// Clock is the rate of the... in hertz.
	// If zero, automatic detection attempted...
	Clock int

	conn       net.Conn
	sequence   uint16
	timestamp  uint32
	previous   time.Time
	syncSource uint32
}

func (s *Session) init() {
	s.sequence = uint16(rand.Intn(0xffff)) // max uint16
	s.syncSource = rand.Uint32()
	s.timestamp = rand.Uint32()
}

// Transmit sends the encoded form of packet to the destination address in s.
// The Session will manage
func (s *Session) Transmit(packet *Packet) error {
	if packet.Header.Version == 0 {
		packet.Header.Version = VersionRFC3550
	}

	packet.Header.Sequence = s.sequence
	s.sequence++

	ticks := ticksSince(s.previous, s.Clock)
	packet.Header.Timestamp = s.timestamp + ticks
	s.previous = time.Now()
	s.timestamp += ticks

	if packet.Header.SyncSource == 0 {
		packet.Header.SyncSource = s.syncSource
	}

	b, err := Marshal(packet)
	if err != nil {
		return fmt.Errorf("marshal packet: %w", err)
	}
	n, err := s.conn.Write(b)
	if n != len(b) {
		if err != nil {
			return fmt.Errorf("short write %d bytes: %w", n, err)
		}
		return fmt.Errorf("short write (%d bytes)", n)
	}
	return err
}

func ticksSince(t time.Time, clockRate int) (ticks uint32) {
	dur := int(time.Since(t)/time.Second) * clockRate
	return uint32(dur)
}
