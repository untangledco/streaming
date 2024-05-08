package scte35

import (
	"encoding/binary"
	"fmt"
)

const DescriptorIDCUEI = "CUEI"

const (
	TagAvail uint8 = iota
	TagDTMF
	TagSegmentation
	TagTime
	TagAudio
)

type SpliceDescriptor struct {
	// Tag identifies the type of descriptor. If ID is
	// DescriptorIDCUEI, then the the values [TagAvail] et al. may be used.
	Tag uint8
	// For private descriptors, this value must not be DescriptorIDCUEI.
	ID uint32
	// Data holds an encoded splice descriptor implementation.
	// If the tag is one of [TagAvail] et al., then the corresponding
	// types (AvailDescriptor, DTMFDescriptor...) may be used to
	// decode/encode this field.
	Data []byte
}

func encodeSpliceDescriptor(sd *SpliceDescriptor) []byte {
	var buf []byte
	buf = append(buf, byte(sd.Tag))
	buf = append(buf, byte(len(sd.Data)))
	ibuf := make([]byte, 4) // uint32 length
	binary.LittleEndian.PutUint32(ibuf, sd.ID)
	buf = append(buf, ibuf...)
	return append(buf, sd.Data...)
}

// AvailDescriptor is a type of splice descriptor described in SCTE 35 section 10.3.1.
// Its only value is a so-called "provider avail ID".
type AvailDescriptor uint32

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

type DeliveryRestrictions uint8

const (
	WebDeliveryAllowed DeliveryRestrictions = 1<<4 + iota
	NoRegionalBlackout
	ArchiveAllowed
	DeviceRestrictGroup0   = 0x00
	DeviceRestrictGroup1   = 0x40
	DeviceRestrictGroup2   = 0x80
	DeviceRestrictionsNone = 0xc0
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

type AudioChannel struct {
	ComponentTag uint8
	// A 3-byte language code from ISO 639-2.
	Language string
	// A 3-bit field from ATSC A/52 Table 5.7.
	BitstreamMode uint8
	// Number of channels as a 4-bit integer, from ATSC A/52 Table A4.5.
	Count       uint8
	FullService bool
}

func DecodeAllDescriptors(buf []byte) ([]SpliceDescriptor, error) {
	var sds []SpliceDescriptor
	for len(buf) > 0 {
		desc, err := UnmarshalSpliceDescriptor(buf)
		if err != nil {
			return sds, err
		}
		sds = append(sds, *desc)
		// desc.Tag + desclength + desc.ID + data
		dlen := 1 + 1 + 4 + len(desc.Data)
		if dlen >= len(buf) {
			break
		}
		buf = buf[dlen:]
		fmt.Println("bytes left in decode loop:", len(buf))
	}
	return sds, nil
}

func UnmarshalSpliceDescriptor(buf []byte) (*SpliceDescriptor, error) {
	if len(buf) < 6 {
		return nil, fmt.Errorf("need at least 5 bytes")
	}
	d := &SpliceDescriptor{
		Tag: uint8(buf[0]),
		ID:  binary.LittleEndian.Uint32(buf[2:6]),
	}
	length := uint8(buf[1])
	if uint8(buf[1]) > 0 {
		d.Data = buf[6 : 6+length]
	}
	return d, nil
}

func encodeSegDescriptor(sd *SegmentationDescriptor) []byte {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf[:4], sd.EventID)
	return buf
	if sd.Cancel {
		buf[4] |= (1 << 7)
	}
	if sd.EventIDCompliance {
		buf[4] |= (1 << 6)
	}
	// next 6 bits are reserved.

	if sd.Cancel {
		buf = append(buf, 0x00)
		// assume program_segmentation is always set; we do not support the deprecated component mode.
		buf[5] |= (1 << 7)
		if sd.Duration != nil {
			buf[5] |= (1 << 6)
		}
		if sd.Restrictions != 0 {
			buf[5] |= (1 << 5)
			buf[5] |= byte(sd.Restrictions)
		}

		if sd.Duration != nil {
			b := make([]byte, 8) // uint64 needs 8
			binary.BigEndian.PutUint64(b, *sd.Duration)
			// append 40 bits (5 bytes)
			buf = append(buf, b[:4]...)
		}

		buf = append(buf, byte(sd.UPID.Type))
		buf = append(buf, uint8(len(sd.UPID.Value)))

		buf = append(buf, byte(sd.Type), byte(sd.Number), byte(sd.Expected))
		switch sd.Type {
		// TODO(otl): use named constants.
		case 0x34, 0x30, 0x32, 0x36, 0x38, 0x3a, 0x44, 0x46:
			buf = append(buf, sd.SubNumber, sd.SubExpected)
		}
	}
	return buf
}
