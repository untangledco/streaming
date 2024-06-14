package mpegts

import (
	"errors"
	"fmt"
	"io"
)

type Scanner struct {
	buf    []byte
	rd     io.Reader
	packet *Packet
	err    error
}

func NewScanner(rd io.Reader) *Scanner {
	return &Scanner{
		buf: make([]byte, PacketSize, PacketSize),
		rd:  rd,
	}
}

func (sc *Scanner) Err() error { return sc.err }

func (sc *Scanner) Packet() *Packet { return sc.packet }

func (sc *Scanner) Scan() bool {
	n, err := sc.rd.Read(sc.buf)
	if n < PacketSize {
		if errors.Is(err, io.EOF) && n == 0 {
			return false
		} else if err != nil {
			sc.err = fmt.Errorf("short read (%d bytes): %w", n, err)
		} else {
			sc.err = fmt.Errorf("short read: read %d bytes", n)
		}
		return false
	}
	if errors.Is(err, io.EOF) && n == PacketSize {
		return false
	} else if errors.Is(err, io.EOF) && n == 0 {
		return false
	} else if err != nil {
		sc.err = err
		return false
	}
	if n == PacketSize {
		p := new(Packet)
		if err := Unmarshal(sc.buf, p); err != nil {
			sc.err = fmt.Errorf("unmarshal: %w", err)
			return false
		}
		sc.packet = p
		return true
	}
	return false
}
