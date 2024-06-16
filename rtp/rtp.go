// Package rtp implements the Real-Time Transport Protocol as
// specified in RFC 3550.
package rtp

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type Packet struct {
	Header  Header
	Payload []byte
}

type Header struct {
	Version uint8
	// TODO(otl): do we store padding bytes? how many?
	padding       bool
	Marker        bool
	Type          PayloadType
	Sequence      uint16
	Timestamp     uint32
	SyncSource    uint32
	ContribSource []uint32
	Extension     *Extension
}

const (
	versionPreSpec uint8 = 0
	VersionDraft   uint8 = 1 << 6
	VersionRFC3550       = 1 << 7
)

type PayloadType uint8

const (
	PayloadMP2T PayloadType = 33
	// ...
)

func (t PayloadType) String() string {
	switch t {
	case PayloadMP2T:
		return "MP2T"
	}
	return "unknown"
}

type Extension struct {
	Profile [2]byte
	Data    []byte
}

var ErrNoPayload = errors.New("no payload")

func Unmarshal(data []byte, p *Packet) error {
	// version (2), padding (1), ext (1), contribcount (4)
	// marker (1), type (7)
	// sequence (16)
	// timestamp (32)
	// syncsource (32)
	// TODO(otl): set this as a constant - headerSize? minPacketSize?
	need := 1 + 1 + 2 + 4 + 4
	if len(data) < need {
		return fmt.Errorf("need %d bytes, have %d", need, len(data))
	}

	p.Header.Version = uint8(data[0] & 0b11000000)
	p.Header.padding = data[0]&0b00100000 > 0
	// extension bit, ignore til later, 0b00010000
	cc := data[0] & 0b00001111
	if cc > 0 {
		p.Header.ContribSource = make([]uint32, cc)
	}

	// m t t t t t t t
	p.Header.Marker = data[1]&0x80 > 0
	p.Header.Type = PayloadType(data[1] & 0x7f)

	p.Header.Sequence = binary.BigEndian.Uint16(data[2:4])
	p.Header.Timestamp = binary.BigEndian.Uint32(data[4:8])
	p.Header.SyncSource = binary.BigEndian.Uint32(data[8:12])

	if len(data) == 12 {
		return ErrNoPayload
	}
	data = data[12:]

	// is the extension bit set?
	if data[0]&0b00010000 > 0 {
		if len(data) < 4 {
			return fmt.Errorf("header extension: %d bytes after header, need %d", len(data), 4)
		}
		ext := &Extension{}
		copy(ext.Profile[:], data[:2])
		length := int(binary.BigEndian.Uint16(data[2:4]))
		if len(data) < length {
			return fmt.Errorf("header extension: reports length %d bytes, only have %d", length, len(data))
		}
		if length > 0 {
			ext.Data = data[4 : 4+length]
			data = data[4+length:]
		} else {
			data = data[4:]
		}
		p.Header.Extension = ext
	}

	need = len(p.Header.ContribSource) * 4 // 32-bit uints * 4 for number of bytes needed
	if len(data) < need {
		return fmt.Errorf("contribution sources: need %d bytes, only have %d", need, len(data))
	}
	var n int
	for i := range p.Header.ContribSource {
		p.Header.ContribSource[i] = binary.BigEndian.Uint32(data[n : n+4])
		n += 4
	}

	if len(data[n:]) == 0 {
		return ErrNoPayload
	}
	p.Payload = data[n:]
	return nil
}

func Marshal(p *Packet) ([]byte, error) {
	if p.Header.Version > VersionRFC3550 {
		return nil, fmt.Errorf("bad version %v", p.Header.Version)
	}
	buf := make([]byte, 12) // see Unmarshal() for size
	buf[0] |= p.Header.Version
	if p.Header.padding {
		buf[0] |= 0b00100000
	}
	if p.Header.Extension != nil {
		buf[0] |= 0b00010000
	}
	maxContribCount := 0x0f // max 4-bit integer
	if len(p.Header.ContribSource) > maxContribCount {
		return nil, fmt.Errorf("contribution source count %d greater than max %d", len(p.Header.ContribSource), maxContribCount)
	}
	buf[0] |= uint8(len(p.Header.ContribSource))

	if p.Header.Marker {
		buf[1] |= 0b10000000
	}
	if p.Header.Type > 0x7f {
		return nil, fmt.Errorf("payload type %s (%d) greater than max %d", p.Header.Type, p.Header.Type, 0x7f)
	}
	buf[1] |= byte(p.Header.Type)

	binary.BigEndian.PutUint16(buf[2:4], p.Header.Sequence)
	binary.BigEndian.PutUint32(buf[4:8], p.Header.Timestamp)
	binary.BigEndian.PutUint32(buf[8:12], p.Header.SyncSource)

	if p.Header.Extension != nil {
		buf = append(buf, p.Header.Extension.Profile[:]...)
		if len(p.Header.Extension.Data) > 0xffff { // max uint16
			return buf, fmt.Errorf("extension data length %d greater than max %d", len(p.Header.Extension.Data), 0xffff)
		}
		buf = binary.BigEndian.AppendUint16(buf, uint16(len(p.Header.Extension.Data)))
		buf = append(buf, p.Header.Extension.Data...)
	}

	for _, src := range p.Header.ContribSource {
		buf = binary.BigEndian.AppendUint32(buf, src)
	}

	return append(buf, p.Payload...), nil
}
