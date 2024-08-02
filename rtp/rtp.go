// Package rtp implements the Real-Time Transport Protocol as
// specified in RFC 3550.
package rtp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
)

// Packet represents a single RTP data packet.
type Packet struct {
	// Header is the RTP fixed header present at the beginning of
	// every packet.
	Header Header
	// Payload contains the raw bytes, excluding the header,
	// transported in a packet.
	Payload []byte
}

// Header represents the "Fixed Header" specified in RFC 3550 section 5.1.
type Header struct {
	// Version specifies the version of RTP used in the Packet.
	// In practice, the only version in use is VersionRFC3550.
	Version uint8

	// TODO(otl): do we store padding bytes? how many?
	padding bool

	// Marker indicates the marker bit is set. The payload type
	// determines how this value is interpreted.
	Marker bool

	// Type specifies the format of the payload transported in the Packet.
	// In general, each type has its own IETF RFC specifying how the payload is encoded.
	// For example, PayloadMP2T is detailed in RFC 2250.
	Type PayloadType

	// Sequence is a monotonically incremented number used by
	// receivers to manage packet loss. The first packet's Sequence
	// should be randomly assigned, then incremented by one for each
	// RTP packet transmitted.
	Sequence uint16

	// Timestamp is the instant sampled of the first byte of the packet.
	// The first packet in a session should have a randomly assigned
	// timestamp. Subsequent timestamps are calculated according to a
	// monotonically incrementing clock. The clock frequency, and how the
	// timestamp should be interpreted, is dictated by the payload type. For
	// instance, the Timestamp field of RTP packets with MPEG payloads
	// represents the number of ticks of a 90KHz clock. Timestamps of GSM
	// audio RTP packets represent ticks of a 8KHz clock.
	Timestamp uint32

	// SyncSource identifies the synchronisation source of the RTP
	// session. It should be randomly assigned at the start of a
	// session and remain unchanged throughout to prevent
	// collisions with other sessions.
	SyncSource uint32

	// ContribSource lists a maximum of 15 contribution sources
	// used to generate the payload. For example, a RTP session for
	// audio transport may list each SyncSource in ContribSource.
	ContribSource []uint32

	// Extension is an optional field which may be used by certain
	// payloads to transmit extra information. The RTP specification
	// discourages the use of Extension. Instead it recommendeds to
	// store extra information in leading bytes of the payload.
	Extension *Extension
}

const (
	versionPreSpec uint8 = 0
	VersionDraft   uint8 = 1 << 6
	VersionRFC3550       = 1 << 7
)

type PayloadType uint8

const (
	PayloadL16Stereo  PayloadType = 10
	PayloadL16Mono PayloadType = 11
	PayloadMP2T PayloadType = 33
	// ...
)

// DynamicPayloadType returns a randomly generated PayloadType from
// the range of allowed values for payloads with non-static PayloadType
// values. For example, transporting text and JPEG XS with RTP requires
// the use of a dynamic payload type.
func DynamicPayloadType() PayloadType {
	floor := 96
	ceil := 127
	return PayloadType(floor + rand.Intn(ceil-floor))
}

func (t PayloadType) String() string {
	if t >= 96 && t <= 127 {
		return "dynamic"
	}
	switch t {
	case PayloadMP2T:
		return "MP2T"
	}
	return "unknown"
}

const (
	ClockMP2T = 90000 // 90KHz
	ClockText = 1000  // 1KHz
)

type Extension struct {
	Profile [2]byte
	Data    []byte
}

var ErrNoPayload = errors.New("no payload")

// minHeaderLength is the minimum number of bytes in a packet header.
// It is calculated from the sum of the following components:
// - 1 byte (version, padding, extension, contrib count)
// - 1 byte (marker + type)
// - 2 bytes (sequence)
const minHeaderLength = 12

func Unmarshal(data []byte, p *Packet) error {
	if len(data) < minHeaderLength {
		return fmt.Errorf("need at least %d bytes, have %d", minHeaderLength, len(data))
	}

	p.Header.Version = uint8(data[0] & 0b11000000)
	p.Header.padding = data[0]&0b00100000 > 0
	hasExtension := data[0]&0b00010000 > 0
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

	if len(data) == minHeaderLength {
		return ErrNoPayload
	}
	data = data[12:]

	if hasExtension {
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

	need := len(p.Header.ContribSource) * 4 // uint32 count * 4 for number of bytes needed
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
	buf := make([]byte, minHeaderLength)
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
