// Package mpegts implements encoding and decoding of MPEG-TS packets as specified in ITU-T H.222.0.
package mpegts

import "fmt"

const PacketSize int = 188

const Sync byte = 'G'

type Packet struct {
	// If true, the packet should be discarded.
	Error bool
	// Indicates whether the payload holds the first byte of a particular payload such as ... TODO(otl)
	PayloadStart bool
	// If true, this packet has higher priority than other packets
	// with the same PID.
	Priority bool
	// Identifies the type of the packet.
	PID PacketID
	// Specifies which algorithm, if any, is used to encrypt the payload.
	Scrambling      Scramble
	Continuity      uint8
	Adaptation      *Adaptation
	Payload         []byte
	emptyAdaptation bool
}

type Scramble uint8

const (
	ScrambleNone Scramble = 0
	_
	ScrambleEven Scramble = 0x80
	ScrambleOdd  Scramble = 0xc0
)

type PacketID uint16

const (
	PAT  PacketID = iota // Program association table
	CAT                  // Conditional access table
	TSDT                 // Transport stream description table
	IPMP
	_                          // reserved through to
	PacketNull PacketID = 8191 // max 13-bit integer
)

func (id PacketID) String() string {
	switch id {
	case PAT:
		return "PAT"
	case CAT:
		return "CAT"
	case TSDT:
		return "TSDT"
	case IPMP:
		return "IPMP"
	case PacketNull:
		return "null"
	}
	return fmt.Sprintf("%d", id)
}

// Adaptation represents an adaptation field including the header.
type Adaptation struct {
	// If true, there is a discontinuity between this packet and
	// either the continuity conter or PCR.
	Discontinuous bool
	// If true, the packet may be used as an index from which to
	// randomly access other points in the stream.
	RandomAccess bool
	// Whether this stream is high priority.
	Priority           bool
	SpliceCountdownSet bool
	PCR                *PCR
	// Original program clock reference.
	OPCR *PCR
	// The number of packets remaining until a splicing point.
	SpliceCountdown uint8
	// Raw bytes of any application-specific data.
	Private []byte
	// Raw bytes of the adaptation extension field.
	Extension []byte
	// A slice of bytes all with the value 0xff; enough to ensure
	// a packet's length is always PacketSize.
	Stuffing []byte
}

// PCR represents a Program Clock Reference.
type PCR struct {
	// 33-bit integer holding the number of ticks of a 90KHz clock.
	Base uint64
	// 9-bit integer holding 27MHz ticks, intended as a
	// constant addition to Base.
	Extension uint16
}

// Returns the number of ticks of a 27MHz clock in p.
func (p PCR) Ticks() uint64 {
	return p.Base*300 + uint64(p.Extension)
}
