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
	SAPType   SAPType
	Encrypted bool
	Cipher    Cipher
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
	if sis.Cipher > maxCipher {
		return nil, fmt.Errorf("encryption algorithm %d larger than max value %d", sis.Cipher, maxCipher)
	}
	// pack 6-bit cipher into next 6. Keep 1 bit for PTSAdjustment.
	buf[4] |= byte(sis.Cipher) << 1
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
		buf1 = append(buf1, encodeSpliceDescriptor(desc)...)
	}
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(buf1)))

	return binary.BigEndian.AppendUint32(buf, calculateCRC32(buf)), nil
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
		info.Cipher = Cipher(buf[1] & 0b01111110)
	}

	pts := make([]byte, 8)
	pts[0] = buf[1] & (1 << 1)
	pts[1] = buf[2]
	pts[2] = buf[3]
	pts[3] = buf[4]
	pts[4] = buf[5]
	info.PTSAdjustment = binary.BigEndian.Uint64(pts)

	if info.Encrypted {
		info.CWIndex = uint8(buf[6])
	}

	// want left-most 12 bits, remaining is used by command length.
	// TODO(otl): still not getting expected values here;
	// check TestDecodeSpliceInfo
	tier := binary.BigEndian.Uint16([]byte{buf[7], buf[8] & 0xf0})
	info.Tier = tier >> 4

	// 4-bits out of buf[8], then all of buf[9] for a 12-bit integer.
	cmdlen := binary.BigEndian.Uint16([]byte{buf[8] & 0x0f, buf[9]})
	cmd, err := decodeCommand(buf[10 : 10+cmdlen+1])
	if err != nil {
		return nil, fmt.Errorf("decode command: %w", err)
	}
	info.Command = cmd
	buf = buf[10+cmdlen+1:]

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

func decodeCommand(buf []byte) (*Command, error) {
	var cmd Command
	cmd.Type = CommandType(buf[0])
	switch cmd.Type {
	case SpliceNull, BandwidthReservation:
		// nothing to decode
	case TimeSignal:
		// check if time specified flag is set.
		// If so, extract the 33-bit integer timestamp.
		if buf[1]&0x80 == 1<<7 {
			b := make([]byte, 8)
			b[3] = buf[1] & 0x01 // ignoring flag and reserved bits
			b[4] = buf[2]
			b[5] = buf[3]
			b[6] = buf[4]
			b[7] = buf[5]
			t := binary.BigEndian.Uint64(b)
			cmd.TimeSignal = &t
		}
	case SpliceInsert:
		var ins Insert
		ins.ID = binary.BigEndian.Uint32(buf[1:5])
		if buf[5]&0x80 > 0 {
			ins.Cancel = true
			cmd.Insert = &ins
			// rebelelder told us to do this.
			return &cmd, nil
		}
		if buf[6]&(1<<7) > 0 {
			ins.OutOfNetwork = true
		}

		// assume program_splice is set at bit 6;

		var durflag bool
		if buf[6]&(1<<5) > 0 {
			durflag = true
		}
		// we don't support deprecated component mode.
		if buf[6]&(1<<4) > 0 {
			ins.Immediate = true
		}
		if buf[6]&(1<<3) > 0 {
			ins.EventIDCompliance = true
		}
		// next 3 bits are reserved.

		if !ins.Immediate {
			// is time_specified_flag set? if so, read the 33-bit time.
			if buf[7]&(1<<7) > 0 {
				b := make([]byte, 3)
				b = append(b, buf[7]&0x01)  // skip reserved bits.
				b = append(b, buf[8:12]...) // read remaining 32 bits.
				dur := binary.BigEndian.Uint64(b)
				ins.SpliceTime = newuint64(dur)
				buf = buf[12:]
			} else {
				buf = buf[8:]
			}
		}

		if durflag {
			a := [5]byte{buf[0], buf[1], buf[2], buf[3], buf[4]}
			ins.Duration = readBreakDuration(a)
			buf = buf[5:]
		}

		ins.ProgramID = binary.BigEndian.Uint16([]byte{buf[0], buf[1]})
		ins.AvailNum = uint8(buf[2])
		ins.AvailExpected = uint8(buf[3])
		cmd.Insert = &ins
	default:
		return nil, fmt.Errorf("TODO: cannot decode command type %s", cmd.Type)
	}
	return &cmd, nil
}
