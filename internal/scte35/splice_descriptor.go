package scte35

import (
	"encoding/binary"
	"fmt"
)

const DescriptorIDCUEI = "CUEI"

// DescriptorIDCUEI in ASCII
const descriptorIDCUEI uint32 = 0x43554549

const (
	TagAvail uint8 = iota
	TagDTMF
	TagSegmentation
	TagTime
	TagAudio
)

type SpliceDescriptor interface {
	// Tag identifies the type of descriptor. If ID is
	// DescriptorIDCUEI, then the the values [TagAvail] et al. may be used.
	Tag() uint8
	// For private descriptors, this value must not be DescriptorIDCUEI.
	ID() uint32
	// Data is the encoded splice descriptor implementation.
	// If Tag is one of [TagAvail] et al., then the corresponding
	// types (AvailDescriptor, DTMFDescriptor...) may be used to
	// decode/encode this field.
	Data() []byte
}

func encodeSpliceDescriptor(sd SpliceDescriptor) []byte {
	var buf []byte
	buf = append(buf, byte(sd.Tag()))
	buf = append(buf, byte(len(sd.Data())))
	buf = binary.LittleEndian.AppendUint32(buf, sd.ID())
	return append(buf, sd.Data()...)
}

// AvailDescriptor is a type of splice descriptor described in SCTE 35 section 10.3.1.
// Its only value is a so-called "provider avail ID".
type AvailDescriptor uint32

func (d AvailDescriptor) Tag() uint8 { return TagAvail }
func (d AvailDescriptor) ID() uint32 { return descriptorIDCUEI }

func (d AvailDescriptor) Data() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(d))
	return buf
}

// DTMFDescriptor is a type of a splice descriptor as described in SCTE 35 10.3.2.
// DTMF stands for [Dual-tone multi-frequency signaling].
//
// [Dual-tone multi-frequency signaling]: https://en.wikipedia.org/wiki/DTMF
type DTMFDescriptor struct {
	Preroll uint8
	// Chars holds a DTMF sequence whose values may only
	// consist of the ASCII values of '0' through '9', '*', and '#'.
	Chars []byte
}

func (d DTMFDescriptor) Tag() uint8 { return TagDTMF }
func (d DTMFDescriptor) ID() uint32 { return descriptorIDCUEI }

func (d DTMFDescriptor) Data() []byte {
	// preroll + char count + chars
	b := make([]byte, 1+1+len(d.Chars))
	b[0] = byte(d.Preroll)
	// set 3 bits, right-most 5 are reserved.
	b[1] = byte(len(d.Chars)) << 5
	i := 2
	for j := range d.Chars {
		b[i] = d.Chars[j]
		i++
	}
	return b
}

type DeliveryRestrictions uint8

const (
	WebDeliveryAllowed DeliveryRestrictions = 1<<4 - iota
	NoRegionalBlackout
	ArchiveAllowed
	DeviceRestrictGroup0   = 0x00
	DeviceRestrictGroup1   = 0x01
	DeviceRestrictGroup2   = 0x02
	DeviceRestrictionsNone = 0x03
)

// SegmentationDescriptor represents the segmentation_descriptor
// structure defined in SCTE 35 section 10.3.3.
type SegmentationDescriptor struct {
	EventID           uint32
	Cancel            bool
	EventIDCompliance bool
	Restrictions      DeliveryRestrictions
	// 40-bit integer representing the number of ticks of a 90KHz clock.
	Duration *uint64
	UPID     UPID
	// Valid types are specified in Table 23, SCTE 35 section 10.3.3.1.
	Type uint8
	// The numbered index of this descriptor in a collection of descriptors.
	Number uint8
	// Expected count of descriptors.
	Expected uint8
	// Numbered index of any subsegment of this descriptor.
	SubNumber uint8
	// Expected count of subsegments.
	SubExpected uint8
}

func (d SegmentationDescriptor) Tag() uint8 { return TagSegmentation }
func (d SegmentationDescriptor) ID() uint32 { return descriptorIDCUEI }

func (d SegmentationDescriptor) Data() []byte {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf[:4], d.EventID)
	if d.Cancel {
		buf[4] |= (1 << 7)
	}
	if d.EventIDCompliance {
		buf[4] |= (1 << 6)
	}
	// next 6 bits are reserved.

	if !d.Cancel {
		buf = append(buf, 0x00)
		// assume program_segmentation is always set; we do not support the deprecated component mode.
		buf[5] |= (1 << 7)
		if d.Duration != nil {
			buf[5] |= (1 << 6)
		}
		if d.Restrictions != 0 {
			buf[5] |= (1 << 5)
			buf[5] |= byte(d.Restrictions)
		}

		if d.Duration != nil {
			b := make([]byte, 8) // uint64 needs 8
			binary.BigEndian.PutUint64(b, *d.Duration)
			// append 40 bits (5 bytes)
			buf = append(buf, b[:4]...)
		}

		buf = append(buf, byte(d.UPID.Type))
		buf = append(buf, uint8(len(d.UPID.Value)))
		buf = append(buf, d.UPID.Value...)

		buf = append(buf, byte(d.Type), byte(d.Number), byte(d.Expected))
		switch d.Type {
		// TODO(otl): use named constants from section 10.3.3.1 Table 23 - segmentation_type_id
		case 0x34, 0x30, 0x32, 0x36, 0x38, 0x3a, 0x44, 0x46:
			buf = append(buf, d.SubNumber, d.SubExpected)
		}
	}
	return buf
}

func unmarshalSegDescriptor(buf []byte) SegmentationDescriptor {
	var desc SegmentationDescriptor
	desc.EventID = binary.BigEndian.Uint32(buf[:4])
	if buf[4]&0b10000000 > 0 {
		desc.Cancel = true
	}
	if buf[4]&0b01000000 > 0 {
		desc.EventIDCompliance = true
	}
	// next 6 bits are reserved

	// always assume program_segmentation_flag is set at 0b10000000
	// we don't support the deprecated component mode.

	if !desc.Cancel {
		// left-most 2 bits are flags for later.
		desc.Restrictions = DeliveryRestrictions(buf[5] & 0b00111111)

		// is segmentation duration flag set?
		if buf[5]&0b01000000 > 0 {
			b := make([]byte, 3)
			b = append(b, buf[6:11]...)
			dur := binary.BigEndian.Uint64(b)
			desc.Duration = &dur
			buf = buf[11:]
		} else {
			buf = buf[6:]
		}

		uplen := int(buf[1])
		desc.UPID = UPID{
			Type:  UPIDType(buf[0]),
			Value: buf[2 : 2+uplen],
		}
		buf = buf[2+uplen:]

		// TODO(otl): use named constants from section 10.3.3.1 Table 23 - segmentation_type_id
		desc.Type = uint8(buf[0])
		desc.Number = uint8(buf[1])
		desc.Expected = uint8(buf[2])
		switch desc.Type {
		// TODO(otl): use named constants from section 10.3.3.1 Table 23 - segmentation_type_id
		case 0x34, 0x30, 0x32, 0x36, 0x38, 0x3a, 0x44, 0x46:
			if desc.Expected > 0 {
				desc.SubNumber = uint8(buf[3])
				desc.SubExpected = uint8(buf[4])
			}
		}
	}

	return desc
}

// UPID represents a segmentation_upid structure as specified in SCTE 35 section 10.3.3.1.
type UPID struct {
	Type UPIDType
	// Value holds the corresponding encoded contents for this UPID's Type.
	// Possible values are given in Table 22 of section 10.3.3.1.
	Value []byte
}

// UPIDType represents a Segmentation UPID type as defined in SCTE 35 section 10.3.3.1.
type UPIDType uint8

// Valid UPIDType values defined in Table 22, SCTE 35 section 10.3.3.1.
const (
	UPIDTypeNone UPIDType = 0 + iota
	_                     // User Defined, deprecated, use MPU.
	_                     // ISCI, deprecated, use AdID.
	UPIDTypeAdID
	UPIDTypeUMID
	_ // ISAN, deprecated, use ISAN.
	UPIDTypeISAN
	UPIDTypeTID
	UPIDTypeTI
	UPIDTypeADI
	UPIDTypeEIDR
	UPIDTypeATSCContentID
	UPIDTypeMPU
	UPIDTypeMID
	UPIDTypeADSInfo
	UPIDTypeURI
	UPIDTypeUUID
	UPIDTypeSCR
	UPIDTypeReserved
)

// TimeDescriptor represents a moment in time as used in the Precision
// Time Protocol (PTP). PTP uses International Atomic Time (TAI) rather
// than UTC time as in NTP.
type TimeDescriptor struct {
	// A 48-bit integer of the number of seconds since the Unix
	// epoch according to TAI.
	Seconds uint64
	// Number of nanoseconds...
	Nanoseconds uint32
	// The current number of seconds between NTP time and
	// TAI for a single instance of time.
	UTCOffset uint16
}

func (d TimeDescriptor) Tag() uint8 { return TagTime }
func (d TimeDescriptor) ID() uint32 { return descriptorIDCUEI }

func (d TimeDescriptor) Data() []byte {
	// 48 bits + 32 bits + 16 bits
	b := make([]byte, 0, 6+4+2)
	b = binary.BigEndian.AppendUint64(b, d.Seconds)
	b = b[:6] // only want 48-bits
	b = binary.BigEndian.AppendUint32(b, d.Nanoseconds)
	return binary.BigEndian.AppendUint16(b, d.UTCOffset)
}

type AudioChannel struct {
	ComponentTag uint8
	// A language code from ISO 639-2.
	Language [3]byte
	// A 3-bit integer from ATSC A/52 Table 5.7.
	BitstreamMode uint8
	// Number of channels as a 4-bit integer, from ATSC A/52 Table A4.5.
	Count       uint8
	FullService bool
}

type AudioDescriptor []AudioChannel

func (d AudioDescriptor) Tag() uint8 { return TagAudio }
func (d AudioDescriptor) ID() uint32 { return descriptorIDCUEI }

func (d AudioDescriptor) Data() []byte {
	var b []byte
	count := len(d)
	b = append(b, byte(count<<4)) // right-most 4 bits are reserved
	for _, ch := range d {
		b = append(b, ch.ComponentTag)
		b = append(b, ch.Language[:]...)
		var c byte
		c |= (ch.BitstreamMode << 5) // set left 3 bits
		c |= (ch.Count & 0x0f)       // only want 4 bits
		if ch.FullService {
			c |= 0x01 // set last remaining bit
		}
		b = append(b, c)
	}
	return b
}

func DecodeAllDescriptors(buf []byte) ([]SpliceDescriptor, error) {
	var sds []SpliceDescriptor
	for len(buf) >= 6 {
		// first byte is tag, second is length of next descriptor.
		dlen := uint8(buf[1])
		desc, err := UnmarshalSpliceDescriptor(buf[:2+dlen])
		if err != nil {
			return sds, err
		}
		sds = append(sds, desc)
		if int(dlen) >= len(buf) {
			break
		}
		buf = buf[2+dlen:]
	}
	return sds, nil
}

// UnmarshalSpliceDescriptor reads exactly one descriptor from buf.
func UnmarshalSpliceDescriptor(buf []byte) (SpliceDescriptor, error) {
	if len(buf) < 6 {
		return nil, fmt.Errorf("short slice: need at least 5 bytes")
	}
	tag := uint8(buf[0])
	length := uint8(buf[1])
	if len(buf[2:]) != int(length) {
		return nil, fmt.Errorf("need %d bytes, have %d", int(length), len(buf[2:]))
	}
	buf = buf[2 : 2+length]
	id := binary.BigEndian.Uint32(buf[:4])
	buf = buf[4:]
	if id != descriptorIDCUEI {
		return PrivateDescriptor{tag, id, buf}, nil
	}
	switch tag {
	case TagAvail:
		return AvailDescriptor(binary.BigEndian.Uint32(buf)), nil
	case TagSegmentation:
		return unmarshalSegDescriptor(buf), nil
	}
	return nil, fmt.Errorf("unmarshal %d unsupported", tag)
}

type PrivateDescriptor struct {
	PTag  uint8
	PID   uint32
	PData []byte
}

func (d PrivateDescriptor) Tag() uint8   { return d.PTag }
func (d PrivateDescriptor) ID() uint32   { return d.PID }
func (d PrivateDescriptor) Data() []byte { return d.PData }
