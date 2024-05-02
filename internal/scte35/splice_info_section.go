// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or   implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package scte35

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bamiaux/iobit"
)

const (
	// TableID is an 8-bit field that shall be 0xFC.
	TableID = 0xFC
	// SectionSyntaxIndicator is a 1-bit field that should always be set to '0'.
	SectionSyntaxIndicator = false
	// PrivateIndicator is a 1-bit flag that shall be set to 0.
	PrivateIndicator = false
)

const (
	// SAPType1 indicates closed GOP with no leading pictures.
	SAPType1 uint32 = iota
	// SAPType2 indicates closed GOP with leading pictures.
	SAPType2
	// SAPType3 indicates Open GOP.
	SAPType3
	// SAPTypeNotSpecified indicates the type of SAP, if any, is not signaled.
	SAPTypeNotSpecified
)

// SpliceInfoSection shall be carried in transport packets whereby only one
// section or partial section may be in any transport packet.
// Splice_info_sections shall always start at the beginning of a transport
// packet payload. When a section begins in a transport packet and this is the
// first packet of the splice_info_section, the pointer_field shall be present
// and equal to 0x00 and the payload_unit_start_indicator bit shall be equal to
// one (per the requirements of section syntax usage per [MPEG Systems]).
type SpliceInfoSection struct {
	EncryptedPacket     EncryptedPacket
	SpliceCommand       SpliceCommand
	SpliceDescriptors   SpliceDescriptors
	SAPType             uint32
	PreRollMilliSeconds uint32 // no corresponding binary field
	PTSAdjustment       uint64
	ProtocolVersion     uint32
	Tier                uint32
	alignmentStuffing   []byte // alignment_stuffing
	ecrc32              []byte // decoded e_crc_32
	crc32               []byte // decoded crc_32
}

// Base64 returns the SpliceInfoSection as a base64 encoded string.
func (sis *SpliceInfoSection) Base64() string {
	b, err := sis.Encode()
	if err != nil {
		Logger.Printf("Error encoding splice_info_section: %s\n", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

// Decode the contents of a byte array into this SpliceInfoSection.
func (sis *SpliceInfoSection) Decode(b []byte) (err error) {
	r := iobit.NewReader(b)
	r.Skip(8) // table_id (shall be 0xFC)
	r.Skip(1) // section_syntax_indicator (shall be 0)
	r.Skip(1) // private_indicator (shall be 0)
	sis.SAPType = r.Uint32(2)
	r.Skip(12) // section_length (informative, can ignore)

	sis.ProtocolVersion = r.Uint32(8)
	encryptedPacket := r.Bit()
	sis.EncryptedPacket.EncryptionAlgorithm = r.Uint32(6)
	sis.PTSAdjustment = r.Uint64(33)
	sis.EncryptedPacket.CWIndex = r.Uint32(8)
	sis.Tier = r.Uint32(12)

	spliceCommandLength := int(r.Uint32(12)) // in bytes
	spliceCommandType := r.Uint32(8)
	switch spliceCommandLength {
	case 0xFFF:
		// legacy signal, decode and skip (buffer underflow expected here)
		r2 := r.Peek()
		sis.SpliceCommand, err = decodeSpliceCommand(spliceCommandType, r2.LeftBytes())
		if err != nil && !errors.Is(err, ErrBufferUnderflow) {
			return err
		}
		r.Skip(uint(sis.SpliceCommand.length() * 8))
	default:
		// standard signal, decode as usual
		sis.SpliceCommand, err = decodeSpliceCommand(spliceCommandType, r.Bytes(spliceCommandLength))
		if err != nil {
			return err
		}
	}

	descriptorLoopLength := int(r.Uint32(16)) // bytes
	sis.SpliceDescriptors, err = decodeSpliceDescriptors(r.Bytes(descriptorLoopLength))
	if err != nil {
		return err
	}

	if encryptedPacket {
		stuffedBytes := (int(r.LeftBits()) - 64) / 8
		if stuffedBytes > 0 {
			sis.alignmentStuffing = r.Bytes(stuffedBytes)
		}
		sis.ecrc32 = r.Bytes(4)
	} else {
		stuffedBytes := (int(r.LeftBits()) - 32) / 8
		if stuffedBytes > 0 {
			sis.alignmentStuffing = r.Bytes(stuffedBytes)
		}
	}
	sis.crc32 = r.Bytes(4)

	if err := readerError(r); err != nil {
		return fmt.Errorf("splice_info_section: %w", err)
	}

	if err = verifyCRC32(b); err != nil {
		return fmt.Errorf("splice_info_section: %w", err)
	}

	return nil
}

// Duration attempts to return the duration of the signal.
func (sis *SpliceInfoSection) Duration() time.Duration {
	// if this is a splice insert with a duration, use it
	if sc, ok := sis.SpliceCommand.(*SpliceInsert); ok {
		if sc.BreakDuration != nil {
			return TicksToDuration(sc.BreakDuration.Duration)
		}
	}

	ticks := uint64(0)
	for _, sd := range sis.SpliceDescriptors {
		if sdt, ok := sd.(*SegmentationDescriptor); ok {
			if sdt.SegmentationDuration != nil {
				ticks += *sdt.SegmentationDuration
			}
		}
	}
	return TicksToDuration(ticks)
}

// Encode returns the binary representation of this SpliceInfoSection as a
// byte array.
func (sis *SpliceInfoSection) Encode() ([]byte, error) {
	buf := make([]byte, sis.length())

	iow := iobit.NewWriter(buf)
	iow.PutUint32(8, TableID)
	iow.PutBit(SectionSyntaxIndicator)
	iow.PutBit(PrivateIndicator)
	iow.PutUint32(2, sis.SAPType)
	iow.PutUint32(12, uint32(sis.sectionLength()))
	iow.PutUint32(8, sis.ProtocolVersion)

	iow.PutBit(sis.EncryptedPacketFlag())
	iow.PutUint32(6, sis.EncryptedPacket.EncryptionAlgorithm)
	iow.PutUint64(33, sis.PTSAdjustment)
	iow.PutUint32(8, sis.EncryptedPacket.CWIndex)
	iow.PutUint32(12, sis.Tier)

	if sis.SpliceCommand != nil {
		iow.PutUint32(12, uint32(sis.SpliceCommand.length()))
		iow.PutUint32(8, sis.SpliceCommand.Type())
		sc, err := sis.SpliceCommand.encode()
		if err != nil {
			return buf, err
		}
		if _, err = iow.Write(sc); err != nil {
			return buf, err
		}
	}

	iow.PutUint32(16, uint32(sis.descriptorLoopLength()))
	for _, sd := range sis.SpliceDescriptors {
		sde, err := sd.encode()
		if err != nil {
			return buf, err
		}
		if _, err = iow.Write(sde); err != nil {
			return buf, err
		}
	}

	// alignment_stuffing
	_, _ = iow.Write(sis.alignmentStuffing)

	// Encoding encrypted signals is untested.
	if sis.EncryptedPacket.EncryptionAlgorithm != EncryptionAlgorithmNone {
		iow.PutUint32(32, calculateCRC32(buf[:iow.Index()/8])) // E_CRC_32
	}

	// Re-calculate CRC_32 to ensure correctness
	iow.PutUint32(32, calculateCRC32(buf[:iow.Index()/8])) // CRC_32

	return buf, iow.Flush()
}

// EncryptedPacketFlag returns the value of encrypted_packet_flag
func (sis *SpliceInfoSection) EncryptedPacketFlag() bool {
	return sis.EncryptedPacket.EncryptionAlgorithm != EncryptionAlgorithmNone
}

// Hex returns the SpliceInfoSection as a hexadecimal encoded string.
func (sis *SpliceInfoSection) Hex() string {
	b, err := sis.Encode()
	if err != nil {
		Logger.Printf("Error encoding splice_info_section: %s\n", err)
		return ""
	}
	return hex.EncodeToString(b)
}

// SAPTypeName returns the Stream Access Point type name.
func (sis *SpliceInfoSection) SAPTypeName() string {
	switch sis.SAPType {
	case SAPType1:
		return "Type 1"
	case SAPType2:
		return "Type 2"
	case SAPType3:
		return "Type 3"
	default:
		return "Not Specified"
	}
}

// length returns the expected length of the encoded splice_info_section, in
// bytes.
func (sis *SpliceInfoSection) length() int {
	length := 8                       // table_id
	length++                          // section_syntax_indicator
	length++                          // private_indicator
	length += 2                       // reserved
	length += 12                      // section_length (bytes remaining)
	length += sis.sectionLength() * 8 // everything else (bytes -> bits)
	return length / 8
}

// sectionLength returns the section_length, in bytes
func (sis *SpliceInfoSection) sectionLength() int {
	length := 8  // protocol_version
	length++     // encrypted_packet
	length += 6  // encryption_algorithm
	length += 33 // pts_adjustment
	length += 8  // cw_index
	length += 12 // tier
	length += 12 // splice_command_length
	length += 8  // splice_command_type
	if sis.SpliceCommand != nil {
		length += sis.SpliceCommand.length() * 8 // bytes -> bits
	}
	length += 16                             // descriptor_loop_length (bytes remaining value)
	length += sis.descriptorLoopLength() * 8 // bytes -> bits
	length += len(sis.alignmentStuffing) * 8 // bytes -> bits
	// we don't officially support encrypted signals so this section is untested.
	// It's implemented here as a base-line if/when we decide to support
	// encryption (ie, use at your own risk)
	if sis.EncryptedPacket.EncryptionAlgorithm != EncryptionAlgorithmNone {
		length += 32 // ECRC_32
	}

	length += 32 // CRC_32
	return length / 8
}

// descriptorLoopLength return the descriptor_loop_length
func (sis *SpliceInfoSection) descriptorLoopLength() int {
	length := 0
	for _, d := range sis.SpliceDescriptors {
		length += 8              // splice_descriptor_tag
		length += 8              // descriptor_length
		length += d.length() * 8 // splice_descriptor()
	}
	return length / 8
}

// iSIS is an internal SpliceInfoSection used to support (un)marshalling
// polymorphic fields.
type iSIS struct {
	EncryptedPacket      EncryptedPacket
	SpliceCommandRaw     json.RawMessage
	SpliceNull           *SpliceNull
	SpliceSchedule       *SpliceSchedule
	SpliceInsert         *SpliceInsert
	TimeSignal           *TimeSignal
	BandwidthReservation *BandwidthReservation
	PrivateCommand       *PrivateCommand
	SpliceDescriptors    SpliceDescriptors
	SAPType              *uint32
	PTSAdjustment        uint64
	ProtocolVersion      uint32
	Tier                 uint32
}

// SpliceCommand returns the polymorphic splice_command.
func (i *iSIS) SpliceCommand() SpliceCommand {
	if i.SpliceNull != nil {
		return i.SpliceNull
	}
	if i.SpliceSchedule != nil {
		return i.SpliceSchedule
	}
	if i.SpliceInsert != nil {
		return i.SpliceInsert
	}
	if i.TimeSignal != nil {
		return i.TimeSignal
	}
	if i.BandwidthReservation != nil {
		return i.BandwidthReservation
	}

	// no valid splice_command?
	if i.SpliceCommandRaw == nil {
		return nil
	}

	// struct to determine the splice command's type
	type sctype struct {
		Type uint32 `json:"type"`
	}

	// get the type
	var st sctype
	if err := json.Unmarshal(i.SpliceCommandRaw, &st); err != nil {
		Logger.Printf("error unmarshalling splice command type: %s", err)
		return nil
	}

	// and decode it
	sc := NewSpliceCommand(st.Type)
	if err := json.Unmarshal(i.SpliceCommandRaw, sc); err != nil {
		Logger.Printf("error unmarshalling splice command: %s", err)
		return nil
	}

	return sc
}
