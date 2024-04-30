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
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/bamiaux/iobit"
)

const (
	// SegmentationDescriptorTag is the splice_descriptor_tag for
	// segmentation_descriptor
	SegmentationDescriptorTag = 0x02

	// SegmentationTypeNotIndicated is the segmentation_type_id for Not Indicated.
	SegmentationTypeNotIndicated = 0x00
	// SegmentationTypeContentIdentification is the segmentation_type_id for
	// Content Identification.
	SegmentationTypeContentIdentification = 0x01
	// SegmentationTypeProgramStart is the segmentation_type_id for Program Start.
	SegmentationTypeProgramStart = 0x10
	// SegmentationTypeProgramEnd is the segmentation_type_id for Program End.
	SegmentationTypeProgramEnd = 0x11
	// SegmentationTypeProgramEarlyTermination is the segmentation_type_id for
	// Program Early Termination.
	SegmentationTypeProgramEarlyTermination = 0x12
	// SegmentationTypeProgramBreakaway is the segmentation_type_id for
	// Program Breakaway.
	SegmentationTypeProgramBreakaway = 0x13
	// SegmentationTypeProgramResumption is the segmentation_type_id for Program
	// Resumption.
	SegmentationTypeProgramResumption = 0x14
	// SegmentationTypeProgramRunoverPlanned is the segmentation_type_id for
	// Program Runover Planned.
	SegmentationTypeProgramRunoverPlanned = 0x15
	// SegmentationTypeProgramRunoverUnplanned is the segmentation_type_id for
	// Program Runover Unplanned.
	SegmentationTypeProgramRunoverUnplanned = 0x16
	// SegmentationTypeProgramOverlapStart is the segmentation_type_id for Program
	// Overlap Start.
	SegmentationTypeProgramOverlapStart = 0x17
	// SegmentationTypeProgramBlackoutOverride is the segmentation_type_id for
	// Program Blackout Override.
	SegmentationTypeProgramBlackoutOverride = 0x18
	// SegmentationTypeProgramStartInProgress is the segmentation_type_id for
	// Program Start - In Progress.
	SegmentationTypeProgramStartInProgress = 0x19
	// SegmentationTypeChapterStart is the segmentation_type_id for Chapter Start.
	SegmentationTypeChapterStart = 0x20
	// SegmentationTypeChapterEnd is the segmentation_type_id for Chapter End.
	SegmentationTypeChapterEnd = 0x21
	// SegmentationTypeBreakStart is the segmentation_type_id for Break Start.
	// Added in ANSI/SCTE 2017.
	SegmentationTypeBreakStart = 0x22
	// SegmentationTypeBreakEnd is the segmentation_type_id for Break End.
	// Added in ANSI/SCTE 2017.
	SegmentationTypeBreakEnd = 0x23
	// SegmentationTypeOpeningCreditStart is the segmentation_type_id for
	// Opening Credit Start. Added in ANSI/SCTE 2020.
	SegmentationTypeOpeningCreditStart = 0x24
	// SegmentationTypeOpeningCreditEnd is the segmentation_type_id for
	// Opening Credit End. Added in ANSI/SCTE 2020.
	SegmentationTypeOpeningCreditEnd = 0x25
	// SegmentationTypeClosingCreditStart is the segmentation_type_id for
	// Closing Credit Start. Added in ANSI/SCTE 2020.
	SegmentationTypeClosingCreditStart = 0x26
	// SegmentationTypeClosingCreditEnd is the segmentation_type_id for
	// Closing Credit End. Added in ANSI/SCTE 2020.
	SegmentationTypeClosingCreditEnd = 0x27
	// SegmentationTypeProviderAdStart is the segmentation_type_id for Provider
	// Ad Start.
	SegmentationTypeProviderAdStart = 0x30
	// SegmentationTypeProviderAdEnd is the segmentation_type_id for Provider Ad
	// End.
	SegmentationTypeProviderAdEnd = 0x31
	// SegmentationTypeDistributorAdStart is the segmentation_type_id for
	// Distributor Ad Start.
	SegmentationTypeDistributorAdStart = 0x32
	// SegmentationTypeDistributorAdEnd is the segmentation_type_id for
	// Distributor Ad End.
	SegmentationTypeDistributorAdEnd = 0x33
	// SegmentationTypeProviderPOStart is the segmentation_type_id for Provider
	// PO Start.
	SegmentationTypeProviderPOStart = 0x34
	// SegmentationTypeProviderPOEnd is the segmentation_type_id for Provider PO
	// End.
	SegmentationTypeProviderPOEnd = 0x35
	// SegmentationTypeDistributorPOStart is the segmentation_type_id for
	// Distributor PO Start.
	SegmentationTypeDistributorPOStart = 0x36
	// SegmentationTypeDistributorPOEnd is the segmentation_type_id for
	// Distributor PO End.
	SegmentationTypeDistributorPOEnd = 0x37
	// SegmentationTypeProviderOverlayPOStart is the segmentation_type_id for
	// Provider Overlay Placement Opportunity Start.
	SegmentationTypeProviderOverlayPOStart = 0x38
	// SegmentationTypeProviderOverlayPOEnd is the segmentation_type_id for
	// Provider Overlay Placement Opportunity End.
	SegmentationTypeProviderOverlayPOEnd = 0x39
	// SegmentationTypeDistributorOverlayPOStart is the segmentation_type_id for
	// Distributor Overlay Placement Opportunity Start.
	SegmentationTypeDistributorOverlayPOStart = 0x3a
	// SegmentationTypeDistributorOverlayPOEnd is the segmentation_type_id for
	// Distributor Overlay Placement Opportunity End.
	SegmentationTypeDistributorOverlayPOEnd = 0x3b
	// SegmentationTypeProviderPromoStart is the segmentation_type_id for
	// Provider Promo Start. Added in ANSI/SCTE 2020.
	SegmentationTypeProviderPromoStart = 0x3c
	// SegmentationTypeProviderPromoEnd is the segmentation_type_id for
	// Provider Promo End. Added in ANSI/SCTE 2020.
	SegmentationTypeProviderPromoEnd = 0x3d
	// SegmentationTypeDistributorPromoStart is the segmentation_type_id for
	// Distributor Promo Start. Added in ANSI/SCTE 2020.
	SegmentationTypeDistributorPromoStart = 0x3e
	// SegmentationTypeDistributorPromoEnd is the segmentation_type_id for
	// Distributor Promo End. Added in ANSI/SCTE 2020.
	SegmentationTypeDistributorPromoEnd = 0x3f
	// SegmentationTypeUnscheduledEventStart is the segmentation_type_id for
	// Unscheduled Event Start.
	SegmentationTypeUnscheduledEventStart = 0x40
	// SegmentationTypeUnscheduledEventEnd is the segmentation_type_id for
	// Unscheduled Event End.
	SegmentationTypeUnscheduledEventEnd = 0x41
	// SegmentationTypeAltConOppStart is the segmentation_type_id for
	// Alternate Content Opportunity Start. Added in ANSI/SCTE 2020.
	SegmentationTypeAltConOppStart = 0x42
	// SegmentationTypeAltConOppEnd is the segmentation_type_id for
	// Alternate Content Opportunity End. Added in ANSI/SCTE 2020.
	SegmentationTypeAltConOppEnd = 0x43
	// SegmentationTypeProviderAdBlockStart is the segmentation_type_id for
	// Provider Ad Block Start. Added in ANSI/SCTE 2020.
	SegmentationTypeProviderAdBlockStart = 0x44
	// SegmentationTypeProviderAdBlockEnd is the segmentation_type_id for
	// Provider Ad Block End. Added in ANSI/SCTE 2020.
	SegmentationTypeProviderAdBlockEnd = 0x45
	// SegmentationTypeDistributorAdBlockStart is the segmentation_type_id for
	// Distributor Ad Block Start. Added in ANSI/SCTE 2020.
	SegmentationTypeDistributorAdBlockStart = 0x46
	// SegmentationTypeDistributorAdBlockEnd is the segmentation_type_id for
	// Distributor Ad Block End. Added in ANSI/SCTE 2020.
	SegmentationTypeDistributorAdBlockEnd = 0x47
	// SegmentationTypeNetworkStart is the segmentation_type_id for Network Start.
	SegmentationTypeNetworkStart = 0x50
	// SegmentationTypeNetworkEnd is the segmentation_type_id for Network End.
	SegmentationTypeNetworkEnd = 0x51
)

// SegmentationDescriptor is an implementation of a splice_descriptor(). It
// provides an optional extension to the time_signal() and splice_insert()
// commands that allows for segmentation messages to be sent in a time/video
// accurate method. This descriptor shall only be used with the time_signal(),
// splice_insert() and the splice_null() commands. The time_signal() or
// splice_insert() message should be sent at least once a minimum of 4 seconds
// in advance of the signaled splice_time() to permit the insertion device to
// place the splice_info_section( ) accurately.
type SegmentationDescriptor struct {
	XMLName                          xml.Name                          `xml:"http://www.scte.org/schemas/35 SegmentationDescriptor" json:"-"`
	JSONType                         uint32                            `xml:"-" json:"type"`
	DeliveryRestrictions             *DeliveryRestrictions             `xml:"http://www.scte.org/schemas/35 DeliveryRestrictions" json:"deliveryRestrictions,omitempty"`
	SegmentationUPIDs                []SegmentationUPID                `xml:"http://www.scte.org/schemas/35 SegmentationUpid" json:"segmentationUpids,omitempty"`
	Components                       []SegmentationDescriptorComponent `xml:"http://www.scte.org/schemas/35 Component" json:"components,omitempty"`
	SegmentationEventID              uint32                            `xml:"segmentationEventId,attr,omitempty" json:"segmentationEventId,omitempty"`
	SegmentationEventCancelIndicator bool                              `xml:"segmentationEventCancelIndicator,attr,omitempty" json:"segmentationEventCancelIndicator,omitempty"`
	SegmentationDuration             *uint64                           `xml:"segmentationDuration,attr" json:"segmentationDuration,omitempty"`
	SegmentationTypeID               uint32                            `xml:"segmentationTypeId,attr,omitempty" json:"segmentationTypeId,omitempty"`
	SegmentNum                       uint32                            `xml:"segmentNum,attr,omitempty" json:"segmentNum,omitempty"`
	SegmentsExpected                 uint32                            `xml:"segmentsExpected,attr,omitempty" json:"segmentsExpected,omitempty"`
	SubSegmentNum                    *uint32                           `xml:"subSegmentNum,attr" json:"subSegmentNum,omitempty"`
	SubSegmentsExpected              *uint32                           `xml:"subSegmentsExpected,attr" json:"subSegmentsExpected,omitempty"`
}

// Name returns the human-readable string for the segmentation_type_id.
func (sd *SegmentationDescriptor) Name() string {
	switch sd.SegmentationTypeID {
	case SegmentationTypeNotIndicated:
		return "Not Indicated"
	case SegmentationTypeContentIdentification:
		return "Content Identification"
	case SegmentationTypeProgramStart:
		return "Program Start"
	case SegmentationTypeProgramEnd:
		return "Program End"
	case SegmentationTypeProgramEarlyTermination:
		return "Program Early Termination"
	case SegmentationTypeProgramBreakaway:
		return "Program Breakaway"
	case SegmentationTypeProgramResumption:
		return "Program Resumption"
	case SegmentationTypeProgramRunoverPlanned:
		return "Program Runover Planned"
	case SegmentationTypeProgramRunoverUnplanned:
		return "Program Runover Unplanned"
	case SegmentationTypeProgramOverlapStart:
		return "Program Overlap Start"
	case SegmentationTypeProgramBlackoutOverride:
		return "Program Blackout Override"
	case SegmentationTypeProgramStartInProgress:
		return "Program Start - In Progress"
	case SegmentationTypeChapterStart:
		return "Chapter Start"
	case SegmentationTypeChapterEnd:
		return "Chapter End"
	case SegmentationTypeBreakStart:
		return "Break Start"
	case SegmentationTypeBreakEnd:
		return "Break End"
	case SegmentationTypeOpeningCreditStart:
		return "Opening Credit Start"
	case SegmentationTypeOpeningCreditEnd:
		return "Opening Credit End"
	case SegmentationTypeClosingCreditStart:
		return "Closing Credit Start"
	case SegmentationTypeClosingCreditEnd:
		return "Closing Credit End"
	case SegmentationTypeProviderAdStart:
		return "Provider Advertisement Start"
	case SegmentationTypeProviderAdEnd:
		return "Provider Advertisement End"
	case SegmentationTypeDistributorAdStart:
		return "Distributor Advertisement Start"
	case SegmentationTypeDistributorAdEnd:
		return "Distributor Advertisement End"
	case SegmentationTypeProviderPOStart:
		return "Provider Placement Opportunity Start"
	case SegmentationTypeProviderPOEnd:
		return "Provider Placement Opportunity End"
	case SegmentationTypeDistributorPOStart:
		return "Distributor Placement Opportunity Start"
	case SegmentationTypeDistributorPOEnd:
		return "Distributor Placement Opportunity End"
	case SegmentationTypeProviderOverlayPOStart:
		return "Provider Overlay Placement Opportunity Start"
	case SegmentationTypeProviderOverlayPOEnd:
		return "Provider Overlay Placement Opportunity End"
	case SegmentationTypeDistributorOverlayPOStart:
		return "Distributor Overlay Placement Opportunity Start"
	case SegmentationTypeDistributorOverlayPOEnd:
		return "Distributor Overlay Placement Opportunity End"
	case SegmentationTypeProviderPromoStart:
		return "Provider Promo Start"
	case SegmentationTypeProviderPromoEnd:
		return "Provider Promo End"
	case SegmentationTypeDistributorPromoStart:
		return "Distributor Promo Start"
	case SegmentationTypeDistributorPromoEnd:
		return "Distributor Promo End"
	case SegmentationTypeUnscheduledEventStart:
		return "Unscheduled Event Start"
	case SegmentationTypeUnscheduledEventEnd:
		return "Unscheduled Event End"
	case SegmentationTypeAltConOppStart:
		return "Alternate Content Opportunity Start"
	case SegmentationTypeAltConOppEnd:
		return "Alternate Content Opportunity End"
	case SegmentationTypeProviderAdBlockStart:
		return "Provider Ad Block Start"
	case SegmentationTypeProviderAdBlockEnd:
		return "Provider Ad Block End"
	case SegmentationTypeDistributorAdBlockStart:
		return "Distributor Ad Block Start"
	case SegmentationTypeDistributorAdBlockEnd:
		return "Distributor Ad Block End"
	case SegmentationTypeNetworkStart:
		return "Network Start"
	case SegmentationTypeNetworkEnd:
		return "Network End"
	default:
		return "Unknown"
	}
}

// Tag returns the splice_descriptor_tag.
func (sd *SegmentationDescriptor) Tag() uint32 {
	// ensure JSONType is set
	sd.JSONType = SegmentationDescriptorTag
	return SegmentationDescriptorTag
}

// DeliveryNotRestrictedFlag returns the delivery_not_restricted_flag.
func (sd *SegmentationDescriptor) DeliveryNotRestrictedFlag() bool {
	return sd.DeliveryRestrictions == nil
}

// ProgramSegmentationFlag returns the program_segmentation_flag.
func (sd *SegmentationDescriptor) ProgramSegmentationFlag() bool {
	return len(sd.Components) == 0
}

// SegmentationDurationFlag returns the segmentation_duration_flag.
func (sd *SegmentationDescriptor) SegmentationDurationFlag() bool {
	return sd.SegmentationDuration != nil
}

// SegmentationUpidLength return the segmentation_upid_length
func (sd *SegmentationDescriptor) SegmentationUpidLength() int {
	length := 0
	if len(sd.SegmentationUPIDs) == 1 {
		length += len(sd.SegmentationUPIDs[0].valueBytes()) * 8 // segmentation_upid() (bytes -> bits)
	} else if len(sd.SegmentationUPIDs) > 1 {
		// for MID, include type & length with each contained upid
		for _, upid := range sd.SegmentationUPIDs {
			length += 8                          // segmentation_upid_type
			length += 8                          // segmentation_upid_length
			length += len(upid.valueBytes()) * 8 // segmentation_upid (bytes -> bits)
		}
	}
	return length / 8
}

// decode updates this splice_descriptor from binary.
func (sd *SegmentationDescriptor) decode(b []byte) error {
	var err error

	r := iobit.NewReader(b)
	r.Skip(8)  // splice_descriptor_tag
	r.Skip(8)  // descriptor_length
	r.Skip(32) // identifier
	sd.SegmentationEventID = r.Uint32(32)
	sd.SegmentationEventCancelIndicator = r.Bit()
	r.Skip(7) // reserved

	if !sd.SegmentationEventCancelIndicator {
		programSegmentationFlag := r.Bit()
		segmentationDurationFlag := r.Bit()
		deliveryNotRestrictedFlag := r.Bit()

		if !deliveryNotRestrictedFlag {
			sd.DeliveryRestrictions = &DeliveryRestrictions{}
			sd.DeliveryRestrictions.WebDeliveryAllowedFlag = r.Bit()
			sd.DeliveryRestrictions.NoRegionalBlackoutFlag = r.Bit()
			sd.DeliveryRestrictions.ArchiveAllowedFlag = r.Bit()
			sd.DeliveryRestrictions.DeviceRestrictions = r.Uint32(2)
		} else {
			r.Skip(5) // reserved
		}

		if !programSegmentationFlag {
			componentCount := int(r.Uint32(8))
			sd.Components = make([]SegmentationDescriptorComponent, componentCount)
			for i := 0; i < componentCount; i++ {
				c := SegmentationDescriptorComponent{}
				c.Tag = r.Uint32(8)
				r.Skip(7) // reserved
				c.PTSOffset = r.Uint64(33)
				sd.Components[i] = c
			}
		}

		if segmentationDurationFlag {
			dur := r.Uint64(40)
			sd.SegmentationDuration = &dur
		}

		segmentationUpidType := r.Uint32(8)
		segmentationUpidLength := int(r.Uint32(8))
		if segmentationUpidLength > 0 {
			segmentationUpidValue := r.Bytes(segmentationUpidLength)

			if segmentationUpidType == SegmentationUPIDTypeMID {
				upidr := iobit.NewReader(segmentationUpidValue)
				sd.SegmentationUPIDs = []SegmentationUPID{}
				for upidr.LeftBits() > 0 {
					upidType := upidr.Uint32(8)
					upidLength := int(upidr.Uint32(8))
					upidValue := upidr.Bytes(upidLength)
					if len(upidValue) < upidLength {
						Logger.Printf("Cannot read value for segmentation_upid_type %d; %d of %d bytes remaining.", upidType, len(upidValue), upidLength)
					}
					sd.SegmentationUPIDs = append(
						sd.SegmentationUPIDs,
						NewSegmentationUPID(upidType, upidValue),
					)
				}
			} else {
				sd.SegmentationUPIDs = []SegmentationUPID{
					NewSegmentationUPID(segmentationUpidType, segmentationUpidValue),
				}
			}
		}

		sd.SegmentationTypeID = r.Uint32(8)
		sd.SegmentNum = r.Uint32(8)
		sd.SegmentsExpected = r.Uint32(8)

		// these fields are new in 2016 so we need a secondary check whether
		// they were actually included in the binary payload
		if sd.SegmentationTypeID == SegmentationTypeProviderPOStart || sd.SegmentationTypeID == SegmentationTypeDistributorPOStart {
			if r.LeftBits() == 16 {
				n := r.Uint32(8)
				e := r.Uint32(8)
				sd.SubSegmentNum = &n
				sd.SubSegmentsExpected = &e
			}
		}
	}

	if err != nil {
		return err
	}
	if err := readerError(r); err != nil {
		return fmt.Errorf("segmentation_descriptor: %w", err)
	}
	return nil
}

// encode this splice_descriptor to binary.
func (sd *SegmentationDescriptor) encode() ([]byte, error) {
	length := sd.length()

	// add 2 bytes to contain splice_descriptor_tag & descriptor_length
	buf := make([]byte, length+2)
	iow := iobit.NewWriter(buf)
	iow.PutUint32(8, SegmentationDescriptorTag)
	iow.PutUint32(8, uint32(length))
	iow.PutUint32(32, CUEIdentifier)
	iow.PutUint32(32, sd.SegmentationEventID)
	iow.PutBit(sd.SegmentationEventCancelIndicator)
	iow.PutUint32(7, Reserved)

	if !sd.SegmentationEventCancelIndicator {
		iow.PutBit(sd.ProgramSegmentationFlag())
		iow.PutBit(sd.SegmentationDurationFlag())

		iow.PutBit(sd.DeliveryNotRestrictedFlag())
		if sd.DeliveryRestrictions != nil {
			iow.PutBit(sd.DeliveryRestrictions.WebDeliveryAllowedFlag)
			iow.PutBit(sd.DeliveryRestrictions.NoRegionalBlackoutFlag)
			iow.PutBit(sd.DeliveryRestrictions.ArchiveAllowedFlag)
			iow.PutUint32(2, sd.DeliveryRestrictions.DeviceRestrictions)
		} else {
			iow.PutUint32(5, Reserved)
		}

		if !sd.ProgramSegmentationFlag() {
			iow.PutUint32(8, uint32(len(sd.Components)))
			for _, c := range sd.Components {
				iow.PutUint32(8, c.Tag)
				iow.PutUint32(7, Reserved)
				iow.PutUint64(33, c.PTSOffset)
			}
		}

		if sd.SegmentationDurationFlag() {
			iow.PutUint64(40, *sd.SegmentationDuration)
		}

		switch len(sd.SegmentationUPIDs) {
		case 0:
			iow.PutUint32(8, 0x00) // segmentation_upid_type
			iow.PutUint32(8, 0x00) // segmentation_upid_length
		case 1:
			vb := sd.SegmentationUPIDs[0].valueBytes()
			iow.PutUint32(8, sd.SegmentationUPIDs[0].Type)
			iow.PutUint32(8, uint32(len(vb)))
			_, _ = iow.Write(vb)
		default:
			iow.PutUint32(8, SegmentationUPIDTypeMID)
			iow.PutUint32(8, uint32(sd.SegmentationUpidLength()))
			for _, upid := range sd.SegmentationUPIDs {
				vb := upid.valueBytes()
				iow.PutUint32(8, upid.Type)
				iow.PutUint32(8, uint32(len(vb)))
				_, _ = iow.Write(vb)
			}
		}

		iow.PutUint32(8, sd.SegmentationTypeID)
		iow.PutUint32(8, sd.SegmentNum)
		iow.PutUint32(8, sd.SegmentsExpected)

		if sd.SubSegmentNum != nil {
			iow.PutUint32(8, *sd.SubSegmentNum)
		}
		if sd.SubSegmentsExpected != nil {
			iow.PutUint32(8, *sd.SubSegmentsExpected)
		}
	}

	err := iow.Flush()
	return buf, err
}

// descriptorLength returns the descriptor_length
func (sd *SegmentationDescriptor) length() int {
	length := 32 // identifier
	length += 32 // segmentation_event_id
	length++     // segmentation_event_cancel_indicator
	length += 7  // reserved

	// if segmentation_event_cancel_indicator == 0
	if !sd.SegmentationEventCancelIndicator {
		length++    // program_segmentation_flag
		length++    // segmentation_duration_flag
		length++    // delivery_not_restricted_flag
		length += 5 // delivery restriction flags or reserved

		// if program_segmentation_flag == 0
		if !sd.ProgramSegmentationFlag() {
			length += 8 // component_count

			// for i=0 to component_count
			for range sd.Components {
				length += 8  // component_tag
				length += 7  // reserved
				length += 33 // pts_offset
			}
		}
		if sd.SegmentationDurationFlag() {
			length += 40 // segmentation_duration
		}
		length += 8                               // segmentation_upid_type
		length += 8                               // segmentation_upid_length
		length += sd.SegmentationUpidLength() * 8 // segmentation_upid() (bytes -> bits)
		length += 8                               // segmentation_type_id
		length += 8                               // segment_num
		length += 8                               // segments_expected

		if sd.SubSegmentNum != nil {
			length += 8 // sub_segment_num
		}
		if sd.SubSegmentsExpected != nil {
			length += 8 // sub_segments_expected
		}
	}

	return length / 8
}

// table returns the tabular description of this SegmentationDescriptor.
func (sd *SegmentationDescriptor) writeTo(t *table) {
	t.row(0, "segmentation_descriptor() {", nil)
	t.row(1, "splice_descriptor_tag", fmt.Sprintf("%#02x", sd.Tag()))
	t.row(1, "descriptor_length", sd.length())
	t.row(1, "identifier", fmt.Sprintf("%#08x (%s)", CUEIdentifier, CUEIASCII))
	t.row(1, "segmentation_event_id", sd.SegmentationEventID)
	t.row(1, "segmentation_event_cancel_indicator", sd.SegmentationEventCancelIndicator)
	if !sd.SegmentationEventCancelIndicator {
		t.row(1, "program_segmentation_flag", sd.ProgramSegmentationFlag())
		t.row(1, "segmentation_duration_flag", sd.SegmentationDurationFlag())
		t.row(1, "delivery_not_restricted_flag", sd.DeliveryNotRestrictedFlag())
		if sd.DeliveryRestrictions != nil {
			t.row(1, "web_delivery_allowed_flag", sd.DeliveryRestrictions.WebDeliveryAllowedFlag)
			t.row(1, "no_regional_blackout_flag", sd.DeliveryRestrictions.NoRegionalBlackoutFlag)
			t.row(1, "archive_allowed_flag", sd.DeliveryRestrictions.ArchiveAllowedFlag)
			t.row(1, "device_restrictions", fmt.Sprintf("%d (%s)", sd.DeliveryRestrictions.DeviceRestrictions, sd.DeliveryRestrictions.deviceRestrictionsName()))
		}
		if len(sd.Components) > 0 {
			t.row(1, "component_count", len(sd.Components))
			for i, c := range sd.Components {
				t.row(1, "component["+strconv.Itoa(i)+"] {", nil)
				t.row(2, "component_tag", c.Tag)
				t.row(2, "pts_offset", c.PTSOffset)
				t.row(1, "}", nil)
			}
		}
		if sd.SegmentationDurationFlag() {
			t.row(1, "segmentation_duration", sd.SegmentationDuration)
		}

		t.row(1, "segmentation_upid_length", sd.SegmentationUpidLength())
		for i, u := range sd.SegmentationUPIDs {
			t.row(1, "segmentation_upid["+strconv.Itoa(i)+"] {", nil)
			t.row(2, "segmentation_upid_type", fmt.Sprintf("%#02x (%s)", u.Type, u.Name()))
			if u.Type == SegmentationUPIDTypeMPU {
				t.row(2, "format_identifier", u.formatIdentifierString())
			}
			t.row(2, "segmentation_upid", u.Value)
			t.row(1, "}", nil)
		}
	}

	t.row(1, "segmentation_type_id", fmt.Sprintf("%#02x (%s)", sd.SegmentationTypeID, sd.Name()))
	t.row(1, "segment_num", sd.SegmentNum)
	t.row(1, "segments_expected", sd.SegmentsExpected)
	switch sd.SegmentationTypeID {
	case SegmentationTypeProviderPOStart,
		SegmentationTypeDistributorPOStart,
		SegmentationTypeProviderOverlayPOStart,
		SegmentationTypeDistributorOverlayPOStart:
		if sd.SubSegmentNum != nil {
			t.row(1, "sub_segment_num", sd.SubSegmentNum)
		}
		if sd.SubSegmentsExpected != nil {
			t.row(1, "sub_segments_expected", sd.SubSegmentsExpected)
		}
	}
	t.row(0, "}", nil)
}

// SegmentationDescriptorComponent describes the Component element contained
// within the SegmentationDescriptorType XML schema definition.
type SegmentationDescriptorComponent struct {
	Tag       uint32 `xml:"componentTag,attr" json:"componentTag"`
	PTSOffset uint64 `xml:"ptsOffset,attr" json:"ptsOffset"`
}
