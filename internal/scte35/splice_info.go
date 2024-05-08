package scte35

import (
	"encoding/binary"
	"fmt"
)

// SAPType represents the two-bit field used to indicate that a Stream
// Access Point (SAP) in the stream
// as specified in SCTE 35 section 9.6.1.
type SAPType uint8

const (
	SAPClosedGOP SAPType = iota
	SAPClosedGOPLeading
	SAPOpenGOP
	SAPNone
)

func (t SAPType) String() string {
	switch t {
	case SAPClosedGOP:
		return "SAP Type 1 (closed GOP)"
	case SAPClosedGOPLeading:
		return "SAP Type 2 (closed GOP with leading pictures)"
	case SAPOpenGOP:
		return "SAP Type 3 (open GOP)"
	case SAPNone:
		return "none"
	}
	return "invalid"
}

type SpliceInfo struct {
	SAPType             SAPType
	Encrypted           bool
	EncryptionAlgorithm Cipher
	// Holds a 33-bit unsigned integer representing the number of ticks of a 90KHz clock.
	PTSAdjustment uint64
	CWIndex       uint8
	// Holds a 12-bit field representing an authorization tier.
	Tier        uint16
	Command     *Command
	Descriptors []SpliceDescriptor
	CRC32       uint32
}

// fields of Splice Info Section which MUST have their values set...
// as specified in SCTE 35 section 9.6.1.
const (
	tableID          uint8 = 0xfc
	protocolVersion        = 0x0
	sectionSyntax          = false
	privateIndicator       = false
)

// maximum 12-bit uint (2^12 - 1)
const maxTier uint16 = 0xfff

func encodeSpliceInfo(sis *SpliceInfo) ([]byte, error) {
	buf := make([]byte, 5)
	buf[0] = byte(tableID)
	// next 2 bits (section_syntax_indicator, private_indicator) must be 0.
	// 0b00000000
	switch sis.SAPType {
	case SAPClosedGOP:
		// nothing to do
	case SAPClosedGOPLeading:
		buf[0] |= (1 << 2)
	case SAPOpenGOP:
		buf[0] |= (1 << 3)
	case SAPNone:
		buf[0] |= 0b00001100
	default:
		return nil, fmt.Errorf("invalid SAP type %x", sis.SAPType)
	}
	// length, buf[1,2] set at the end
	buf[3] = protocolVersion

	buf = append(buf, 0x00)
	if sis.Encrypted {
		buf[4] |= (1 << 7)
	}
	if sis.EncryptionAlgorithm > maxCipher {
		return nil, fmt.Errorf("encryption algorithm %d larger than max value %d", sis.EncryptionAlgorithm, maxCipher)
	}
	// pack 6-bit cipher into next 6. Keep 1 bit for PTSAdjustment.
	buf[4] |= byte(sis.EncryptionAlgorithm) << 1
	pts := toPTS(sis.PTSAdjustment)
	buf[4] |= pts[0]
	buf = append(buf, pts[1:]...)

	buf = append(buf, byte(sis.CWIndex))

	if sis.Tier > maxTier {
		return nil, fmt.Errorf("tier %d greater than max %d", sis.Tier, maxTier)
	}
	tier := packTier(sis.Tier)
	buf = append(buf, tier[0])
	buf = append(buf, tier[1]<<4)
	// next 4 bits will be from the command length
	if sis.Command == nil {
		return nil, fmt.Errorf("nil command")
	}
	cmd, err := encodeCommand(sis.Command)
	if err != nil {
		return nil, fmt.Errorf("encode splice command: %w", err)
	}
	length := uint16(len(cmd))
	// stuff remaining 4 bits into the last byte.
	buf[len(buf)-1] |= byte(length >> 8)
	buf = append(buf, byte(length))
	buf = append(buf, cmd...)

	var buf1 []byte
	for _, desc := range sis.Descriptors {
		buf1 = append(buf1, encodeSpliceDescriptor(&desc)...)
	}
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(len(buf1)))
	buf = append(buf, b...)

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, calculateCRC32(buf))
	buf = append(buf, b...)
	return buf, nil
}

func packTier(tier uint16) [2]byte {
	var a [2]byte
	// mask off last 4 bits; we want a 12-bit integer.
	a[0] = byte(tier>>8) & 0b00001111
	a[1] = byte(tier)
	return a
}

func decodeSpliceInfo(buf []byte) (*SpliceInfo, error) {
	if len(buf) < 3 {
		return nil, fmt.Errorf("need at least 2 bytes")
	}
	// skip buf[0], we don't store table_id.

	var info SpliceInfo
	// skip 2 bits, straight to sap_type.
	st := buf[1] & 0b00110000
	info.SAPType = SAPType(st >> 4)

	length := binary.BigEndian.Uint16([]byte{buf[1], buf[2]})
	length &= 0x0fff // 12-bit field
	buf = buf[3:]
	if len(buf) != int(length) {
		return nil, fmt.Errorf("message declares %d bytes but have %d", length, len(buf))
	}

	// skip version byte, we don't store version as it's constant.
	if buf[1]&0b10000000 == 1 {
		info.Encrypted = true
		// right-most bit is used by PTSAdjustment.
		info.EncryptionAlgorithm = Cipher(buf[1] & 0b01111110)
	}

	pts := make([]byte, 8)
	pts[0] = buf[1] & (1 << 1)
	pts[1] = buf[2]
	pts[2] = buf[3]
	pts[3] = buf[4]
	pts[4] = buf[5]
	info.PTSAdjustment = binary.BigEndian.Uint64(pts)

	info.CWIndex = uint8(buf[6])

	// want left-most 12 bits, remaining is used by command length.
	// TODO(otl): still not getting expected values here;
	// check TestDecodeSpliceInfo
	tier := binary.BigEndian.Uint16([]byte{buf[7], buf[8] & 0xf0})
	info.Tier = tier

	// 4-bits out of buf[8], then all of buf[9] for a 12-bit integer.
	cmdlength := binary.BigEndian.Uint16([]byte{buf[8] & 0x0f, buf[9]})
	var cmd Command
	cmd.Type = CommandType(buf[10])
	cmdbuf := buf[11 : 11+cmdlength]
	switch cmd.Type {
	case SpliceNull, BandwidthReservation:
		// nothing to decode
	case TimeSignal:
		// check if time specified flag is set.
		if cmdbuf[0]&0x80 == 1<<7 {
			b := make([]byte, 8)
			b[3] = cmdbuf[0] & 0x01 // ignoring flag and reserved bits
			b[4] = cmdbuf[1]
			b[5] = cmdbuf[2]
			b[6] = cmdbuf[3]
			b[7] = cmdbuf[4]
			t := binary.BigEndian.Uint64(b)
			cmd.TimeSignal = &t
		}
	default:
		return nil, fmt.Errorf("cannot decode command type %s", cmd.Type)
	}
	info.Command = &cmd
	buf = buf[11+cmdlength:]

	desclen := binary.BigEndian.Uint16([]byte{buf[0], buf[1]})
	descriptors, err := DecodeAllDescriptors(buf[2 : 2+desclen])
	if err != nil {
		return nil, fmt.Errorf("decode splice descriptors: %w", err)
	}
	info.Descriptors = descriptors

	buf = buf[2+desclen:]
	if info.Encrypted {
		// TODO(otl): handle alignment_stuffing for encrypted packets.
		// skip past E_CRC_32; we don't store it.
		buf = buf[1:]
	}
	info.CRC32 = binary.BigEndian.Uint32(buf)
	return &info, nil
}
