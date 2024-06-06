// Package mpegts implements encoding and decoding of MPEG-TS packets as specified in ITU-T H.222.0.
package mpegts

const PacketSize int = 188

type Packet struct {
	// If true, the packet should be discarded.
	Error bool
	// If true, this packet has higher priority than other packets
	// with the same PID.
	Priority bool
	// Indicates whether the payload holds the first byte of a particular payload such as ... TODO(otl)
	PayloadStart bool
	// Identifies the type of the payload.
	PID uint16
	// Specifies which algorithm, if any, is used to encrypt the payload.
	ScrambleConrol Scramble
	Adaptation     *Adaptation
	Payload        []byte
}

type Scramble uint8

const (
	ScrambleNone Scramble = 0
	_
	ScrambleEven Scramble = 0x80
	ScrambleOdd  Scramble = 0xc0
)

type Adaptation struct{}
