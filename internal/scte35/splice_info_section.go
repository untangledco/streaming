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
	"encoding/xml"
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

	// SAPType1 indicates closed GOP with no leading pictures.
	SAPType1 = uint32(0x0)
	// SAPType2 indicates closed GOP with leading pictures.
	SAPType2 = uint32(0x1)
	// SAPType3 indicates Open GOP.
	SAPType3 = uint32(0x2)
	// SAPTypeNotSpecified indicates the type of SAP, if any, is not signaled.
	SAPTypeNotSpecified = uint32(0x3)
)

// SpliceInfoSection shall be carried in transport packets whereby only one
// section or partial section may be in any transport packet.
// Splice_info_sections shall always start at the beginning of a transport
// packet payload. When a section begins in a transport packet and this is the
// first packet of the splice_info_section, the pointer_field shall be present
// and equal to 0x00 and the payload_unit_start_indicator bit shall be equal to
// one (per the requirements of section syntax usage per [MPEG Systems]).
type SpliceInfoSection struct {
	XMLName             xml.Name          `xml:"http://www.scte.org/schemas/35 SpliceInfoSection"`
	EncryptedPacket     EncryptedPacket   `xml:"http://www.scte.org/schemas/35 EncryptedPacket,omitempty"`
	SpliceCommand       SpliceCommand     `xml:""`
	SpliceDescriptors   SpliceDescriptors `xml:""`
	SAPType             uint32            `xml:"sapType,attr"`
	PreRollMilliSeconds uint32            `xml:"preRollMilliSeconds,attr,omitempty"` // no corresponding binary field
	PTSAdjustment       uint64            `xml:"ptsAdjustment,attr,omitempty"`
	ProtocolVersion     uint32            `xml:"protocolVersion,attr,omitempty"`
	Tier                uint32            `xml:"tier,attr"`
	alignmentStuffing   []byte            // alignment_stuffing
	ecrc32              []byte            // decoded e_crc_32
	crc32               []byte            // decoded crc_32
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

	err := iow.Flush()
	return buf, err
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

// Table returns the tabular description of this SpliceInfoSection as described
// in ANSI/SCTE 35 Table 5.
func (sis *SpliceInfoSection) Table(prefix, indent string) string {
	// top level table is not indented
	t := newTable(prefix, indent)

	t.row(0, "splice_info_section() {", nil)
	t.row(1, "table_id", fmt.Sprintf("%#02x", TableID))
	t.row(1, "section_syntax_indicator", SectionSyntaxIndicator)
	t.row(1, "private_indicator", PrivateIndicator)
	t.row(1, "sap_type", fmt.Sprintf("%d (%s)", sis.SAPType, sis.SAPTypeName()))
	t.row(1, "section_length", sis.sectionLength())
	t.row(0, "}", nil)
	t.row(0, "protocol_version", sis.ProtocolVersion)
	t.row(0, "encryption_algorithm", fmt.Sprintf("%d (%s)", sis.EncryptedPacket.EncryptionAlgorithm, sis.EncryptedPacket.encryptionAlgorithmName()))
	t.row(0, "pts_adjustment", sis.PTSAdjustment)
	t.row(0, "cw_index", sis.EncryptedPacket.CWIndex)
	t.row(0, "tier", sis.Tier)

	if sis.SpliceCommand != nil {
		t.row(0, "splice_command_length", sis.SpliceCommand.length())
		t.row(0, "splice_command_type", fmt.Sprintf("%#02x", sis.SpliceCommand.Type()))
		sis.SpliceCommand.writeTo(t)
	}

	t.row(0, "descriptor_loop_length", sis.descriptorLoopLength())
	for _, sd := range sis.SpliceDescriptors {
		sd.writeTo(t)
	}
	return t.String()
}

// MarshalJSON encodes a SpliceInfoSection to JSON.
func (sis *SpliceInfoSection) MarshalJSON() ([]byte, error) {
	// ensure JSONTypes are all set before marshalling. These are included in
	// each SpliceCommand.Type() and SpliceDescriptor.Tag() implementation.
	sis.SpliceCommand.Type()
	for i := range sis.SpliceDescriptors {
		sis.SpliceDescriptors[i].Tag()
	}

	m := map[string]interface{}{
		"encryptedPacket": sis.EncryptedPacket,
		"spliceCommand":   sis.SpliceCommand,
		"sapType":         sis.SAPType,
		"tier":            sis.Tier,
	}
	if sis.ProtocolVersion > 0 {
		m["protocolVersion"] = sis.ProtocolVersion
	}
	if sis.PTSAdjustment > 0 {
		m["ptsAdjustment"] = sis.PTSAdjustment
	}
	if len(sis.SpliceDescriptors) > 0 {
		m["spliceDescriptors"] = sis.SpliceDescriptors
	}
	return json.Marshal(m)
}

// UnmarshalJSON decodes a SpliceInfoSection from JSON.
func (sis *SpliceInfoSection) UnmarshalJSON(b []byte) error {
	var tmp iSIS
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	sis.EncryptedPacket = tmp.EncryptedPacket
	sis.SpliceCommand = tmp.SpliceCommand()
	sis.SpliceDescriptors = tmp.SpliceDescriptors
	if tmp.SAPType != nil {
		sis.SAPType = *tmp.SAPType
	} else {
		sis.SAPType = SAPTypeNotSpecified
	}
	sis.PTSAdjustment = tmp.PTSAdjustment
	sis.ProtocolVersion = tmp.ProtocolVersion
	sis.Tier = tmp.Tier
	return nil
}

// UnmarshalXML decodes a SpliceInfoSection from XML.
func (sis *SpliceInfoSection) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var tmp iSIS
	if err := d.DecodeElement(&tmp, &start); err != nil {
		return err
	}
	sis.EncryptedPacket = tmp.EncryptedPacket
	sis.SpliceCommand = tmp.SpliceCommand()
	sis.SpliceDescriptors = tmp.SpliceDescriptors
	if tmp.SAPType != nil {
		sis.SAPType = *tmp.SAPType
	} else {
		sis.SAPType = SAPTypeNotSpecified
	}
	sis.PTSAdjustment = tmp.PTSAdjustment
	sis.ProtocolVersion = tmp.ProtocolVersion
	sis.Tier = tmp.Tier
	return nil
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
	EncryptedPacket      EncryptedPacket       `xml:"http://www.scte.org/schemas/35 EncryptedPacket,omitempty" json:"encryptedPacket,omitempty"`
	SpliceCommandRaw     json.RawMessage       `xml:"-" json:"spliceCommand"`
	SpliceNull           *SpliceNull           `xml:"http://www.scte.org/schemas/35 SpliceNull" json:"-"`
	SpliceSchedule       *SpliceSchedule       `xml:"http://www.scte.org/schemas/35 SpliceSchedule" json:"-"`
	SpliceInsert         *SpliceInsert         `xml:"http://www.scte.org/schemas/35 SpliceInsert" json:"-"`
	TimeSignal           *TimeSignal           `xml:"http://www.scte.org/schemas/35 TimeSignal" json:"-"`
	BandwidthReservation *BandwidthReservation `xml:"http://www.scte.org/schemas/35 BandwidthReservation" json:"-"`
	PrivateCommand       *PrivateCommand       `xml:"http://www.scte.org/schemas/35 PrivateCommand" json:"-"`
	SpliceDescriptors    SpliceDescriptors     `xml:",any" json:"spliceDescriptors"`
	SAPType              *uint32               `xml:"sapType,attr" json:"sapType,omitempty"`
	PTSAdjustment        uint64                `xml:"ptsAdjustment,attr" json:"ptsAdjustment"`
	ProtocolVersion      uint32                `xml:"protocolVersion,attr" json:"protocolVersion"`
	Tier                 uint32                `xml:"tier,attr" json:"tier"`
}

// SpliceCommand returns the polymorphic splice_command.
func (i *iSIS) SpliceCommand() SpliceCommand {
	// xml unmarshalls to the corresponding struct
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
