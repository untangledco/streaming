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
	"encoding/binary"
	"testing"
)

/*
func TestDecodeBase64(t *testing.T) {
	// when adding tests that contain multiple splice descriptors, care must be
	// taken to ensure they are in the order specified in the custom UnmarshalXML
	// implementation, otherwise misleading error may occur
	cases := map[string]struct {
		binary   string
		expected SpliceInfoSection
		legacy   bool
	}{
		"Sample 14.1 time_signal - Placement Opportunity Start": {
			binary: "/DA0AAAAAAAA///wBQb+cr0AUAAeAhxDVUVJSAAAjn/PAAGlmbAICAAAAAAsoKGKNAIAmsnRfg==",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &Command{
					Type: CommandTimeSignal,
					TimeSignal: &TimeSignal{
						SpliceTime: SpliceTime{
							PTSTime: uint64ptr(0x072bd0050),
						},
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
						},
						SegmentationEventID:  uint32(0x4800008e),
						SegmentationTypeID:   SegmentationTypeProviderPOStart,
						SegmentationDuration: uint64ptr(0x0001a599b0),
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ca0a18a)),
						},
						SegmentNum: 2,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.2 splice_insert": {
			binary: "/DAvAAAAAAAA///wFAVIAACPf+/+c2nALv4AUsz1AAAAAAAKAAhDVUVJAAABNWLbowo=",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &Command{
					Type: CommandSpliceInsert,
					Insert: &SpliceInsert{
						BreakDuration: &BreakDuration{
							AutoReturn: true,
							Duration:   uint64(0x00052ccf5),
						},
						SpliceEventID:         uint32(0x4800008f),
						OutOfNetworkIndicator: true,
						Program: &SpliceInsertProgram{
							SpliceTime: SpliceTime{
								PTSTime: uint64ptr(0x07369c02e),
							},
						},
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&AvailDescriptor{
						ProviderAvailID: 0x00000135,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.3 time_signal - Placement Opportunity End": {
			binary: "/DAvAAAAAAAA///wBQb+dGKQoAAZAhdDVUVJSAAAjn+fCAgAAAAALKChijUCAKnMZ1g=",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{
						PTSTime: uint64ptr(0x0746290a0),
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x4800008e,
						SegmentationTypeID:  SegmentationTypeProviderPOEnd,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ca0a18a)),
						},
						SegmentNum: 2,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.4 time_signal - Program Start/End": {
			binary: "/DBIAAAAAAAA///wBQb+ek2ItgAyAhdDVUVJSAAAGH+fCAgAAAAALMvDRBEAAAIXQ1VFSUgAABl/nwgIAAAAACyk26AQAACZcuND",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{
						PTSTime: uint64ptr(0x07a4d88b6),
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x48000018,
						SegmentationTypeID:  SegmentationTypeProgramEnd,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ccbc344)),
						},
					},
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x48000019,
						SegmentationTypeID:  SegmentationTypeProgramStart,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ca4dba0)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.5 time_signal - Program Overlap Start": {
			binary: "/DAvAAAAAAAA///wBQb+rr//ZAAZAhdDVUVJSAAACH+fCAgAAAAALKVs9RcAAJUdsKg=",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{
						PTSTime: uint64ptr(0x0aebfff64),
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x48000008,
						SegmentationTypeID:  SegmentationTypeProgramOverlapStart,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ca56cf5)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.6 time_signal - Program Blackout Override / Program End": {
			binary: "/DBIAAAAAAAA///wBQb+ky44CwAyAhdDVUVJSAAACn+fCAgAAAAALKCh4xgAAAIXQ1VFSUgAAAl/nwgIAAAAACygoYoRAAC0IX6w",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{
						PTSTime: uint64ptr(0x0932e380b),
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x4800000a,
						SegmentationTypeID:  SegmentationTypeProgramBlackoutOverride,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ca0a1e3)),
						},
					},
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x48000009,
						SegmentationTypeID:  SegmentationTypeProgramEnd,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ca0a18a)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.7 time_signal - Program End": {
			binary: "/DAvAAAAAAAA///wBQb+rvF8TAAZAhdDVUVJSAAAB3+fCAgAAAAALKVslxEAAMSHai4=",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{
						PTSTime: uint64ptr(0x0aef17c4c),
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x48000007,
						SegmentationTypeID:  SegmentationTypeProgramEnd,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ca56c97)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.8 time_signal - Program Start/End - Placement Opportunity End": {
			binary: "/DBhAAAAAAAA///wBQb+qM1E7QBLAhdDVUVJSAAArX+fCAgAAAAALLLXnTUCAAIXQ1VFSUgAACZ/nwgIAAAAACyy150RAAACF0NVRUlIAAAnf58ICAAAAAAsstezEAAAihiGnw==",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{
						PTSTime: uint64ptr(0x0a8cd44ed),
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x480000ad,
						SegmentationTypeID:  SegmentationTypeProviderPOEnd,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002cb2d79d)),
						},
						SegmentNum: 2,
					},
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x48000026,
						SegmentationTypeID:  SegmentationTypeProgramEnd,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002cb2d79d)),
						},
					},
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
							WebDeliveryAllowedFlag: true,
						},
						SegmentationEventID: 0x48000027,
						SegmentationTypeID:  SegmentationTypeProgramStart,
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002cb2d7b3)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"SpliceInsert With DTMF": {
			binary: "/DAxAAAAAAAAAP/wFAVAAIeuf+/+0AWRK/4AUmXAAC0AfwAMAQpDVUVJUJ81MTkqo5/+gA==",
			expected: SpliceInfoSection{
				SpliceCommand: &SpliceInsert{
					BreakDuration:              &BreakDuration{AutoReturn: true, Duration: 5400000},
					Program:                    NewSpliceInsertProgram(3490025771),
					SpliceEventID:              1073776558,
					SpliceEventCancelIndicator: false,
					SpliceImmediateFlag:        false,
					OutOfNetworkIndicator:      true,
					UniqueProgramID:            45,
					AvailNum:                   0,
					AvailsExpected:             127,
				},
				SpliceDescriptors: []SpliceDescriptor{
					&DTMFDescriptor{
						Preroll:   80,
						DTMFChars: "519*",
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Time Signal with Segmentation Descriptors": {
			binary: "/DBIAAAAAAAAAP/wBQb/tB67hgAyAhdDVUVJQAABEn+fCAgAAAAALzE8BTUAAAIXQ1VFSUAAAEV/nwgIAAAAAC8xPN4jAAAfiOPE",
			expected: SpliceInfoSection{
				SpliceCommand: NewTimeSignal(7316880262),
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: true,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     DeviceRestrictionsNone,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:  8,
								Value: "791755781",
							},
						},
						SegmentationTypeID:  53,
						SegmentationEventID: 1073742098,
					},
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: true,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     DeviceRestrictionsNone,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:  8,
								Value: "791755998",
							},
						},
						SegmentationTypeID:  35,
						SegmentationEventID: 1073741893,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Splice Insert with Avail Descriptor": {
			binary: "/DAqAAAAAAAAAP/wDwUAAHn+f8/+QubGOQAAAAAACgAIQ1VFSQAAAADizteX",
			expected: SpliceInfoSection{
				SpliceCommand: &SpliceInsert{
					Program:               NewSpliceInsertProgram(1122420281),
					SpliceEventID:         31230,
					OutOfNetworkIndicator: true,
				},
				SpliceDescriptors: []SpliceDescriptor{
					&AvailDescriptor{
						ProviderAvailID: 0,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Multiple SegmentationUPIDs": {
			binary: "/DBrAAAAAAAAAP/wBQb/AAAAAABVAlNDVUVJAAAAAn+/DUQKDBR3i+Xj9gAAAAAAAAoMFHeL5eP2AAAAAAAACSZTSUdOQUw6THk5RU1HeEtSMGhGWlV0cE1IZENVVlpuUlVGblp6MTcBA6QTOe8=",
			expected: SpliceInfoSection{
				SpliceCommand: NewTimeSignal(4294967296),
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:  SegmentationUPIDTypeEIDR,
								Value: "10.5239/8BE5-E3F6-0000-0000-0000",
							},
							{
								Type:  SegmentationUPIDTypeEIDR,
								Value: "10.5239/8BE5-E3F6-0000-0000-0000",
							},
							{
								Type:  SegmentationUPIDTypeADI,
								Value: "SIGNAL:Ly9EMGxKR0hFZUtpMHdCUVZnRUFnZz1",
							},
						},
						SegmentationEventID: 2,
						SegmentationTypeID:  SegmentationTypeDistributorPOEnd,
						SegmentNum:          1,
						SegmentsExpected:    3,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Legacy splice_command_length: 0xFFF": {
			binary: "/DA8AAAAAAAAAP///wb+06ACpQAmAiRDVUVJAACcHX//AACky4AMEERJU0NZTVdGMDQ1MjAwMEgxAQEMm4c0",
			expected: SpliceInfoSection{
				SpliceCommand: NewTimeSignal(3550479013),
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:             SegmentationUPIDTypeMPU,
								FormatIdentifier: uint32ptr(1145656131),
								Value:            "WU1XRjA0NTIwMDBI",
							},
						},
						SegmentationDuration: uint64ptr(10800000),
						SegmentationEventID:  39965,
						SegmentationTypeID:   SegmentationTypeProviderAdEnd,
						SegmentNum:           1,
						SegmentsExpected:     1,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
			legacy: true,
		},
		"Signal with non-CUEI descriptor": {
			binary: "/DBPAAAAAAAAAP/wBQb/Gq9LggA5AAVTQVBTCwIwQ1VFSf////9//wAAFI4PDxx1cm46bmJjdW5pLmNvbTpicmM6NDk5ODY2NDM0MQoBbM98zw==",
			expected: SpliceInfoSection{
				SpliceCommand: NewTimeSignal(4742663042),
				SpliceDescriptors: []SpliceDescriptor{
					&PrivateDescriptor{
						Identifier:   1396789331,
						PrivateBytes: []byte{11},
					},
					&SegmentationDescriptor{
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:  SegmentationUPIDTypeURI,
								Value: "urn:nbcuni.com:brc:499866434",
							},
						},
						SegmentationDuration: uint64ptr(1347087),
						SegmentationEventID:  4294967295,
						SegmentationTypeID:   SegmentationTypeProviderAdEnd,
						SegmentNum:           10,
						SegmentsExpected:     1,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Splice Null - Heartbeat": {
			binary: "/DARAAAAAAAAAP/wAAAAAHpPv/8=",
			expected: SpliceInfoSection{
				SpliceCommand: nil,
				Tier:          4095,
				SAPType:       3,
			},
		},
		"Alignment Stuffing without Encryption": {
			binary: "/DAeAAAAAAAAAP///wViAA/nf18ACQAAAAAskJv+YPtE",
			expected: SpliceInfoSection{
				SpliceCommand: &SpliceInsert{
					SpliceEventID:       1644171239,
					Program:             &SpliceInsertProgram{},
					SpliceImmediateFlag: true,
					UniqueProgramID:     9,
				},
				Tier:    4095,
				SAPType: 3,
			},
			legacy: true, // binary wont match because of stuffing
		},
		"UPID with Valid ASCII Invalid UTF8": {
			binary: "/DDHAAAAABc0AP/wBQb/tVo+agCxAhdDVUVJQA4hwH+fCAgAAAAAPj6IcCMAAAIXQ1VFSUAOI1x/nwgIAAAAAD4+iHARAAACF0NVRUlADiHgf58ICAAAAAA+Poi2EAAAAhxDVUVJQA4hyn/fAABSlKwICAAAAAA+Poi2IgAAAkZDVUVJQA4h1n/PAABSlKwNMgoMFHf5uXs0AAAAAAAADhh0eXBlPUxBJmR1cj02MDAwMCZ0aWVy/DDHAAAAAAAAAP8ABQb/HPCt2w==",
			expected: SpliceInfoSection{
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{
						PTSTime: uint64ptr(7337557610),
					},
				},
				SpliceDescriptors: SpliceDescriptors{
					&SegmentationDescriptor{
						SegmentationEventID: uint32(1074667968),
						SegmentationTypeID:  SegmentationTypeBreakEnd,
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: true,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{Type: SegmentationUPIDTypeTI, Value: "1044285552"},
						},
					},
					&SegmentationDescriptor{
						SegmentationEventID: uint32(1074668380),
						SegmentationTypeID:  SegmentationTypeProgramEnd,
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: true,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{Type: SegmentationUPIDTypeTI, Value: "1044285552"},
						},
					},
					&SegmentationDescriptor{
						SegmentationEventID: uint32(1074668000),
						SegmentationTypeID:  SegmentationTypeProgramStart,
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: true,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{Type: SegmentationUPIDTypeTI, Value: "1044285622"},
						},
					},
					&SegmentationDescriptor{
						SegmentationEventID:  uint32(1074667978),
						SegmentationDuration: uint64ptr(5412012),
						SegmentationTypeID:   SegmentationTypeBreakStart,
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: true,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{Type: SegmentationUPIDTypeTI, Value: "1044285622"},
						},
					},
					&SegmentationDescriptor{
						SegmentationEventID:  uint32(1074667990),
						SegmentationTypeID:   0x05,
						SegmentationDuration: uint64ptr(5412012),
						SegmentNum:           6,
						SegmentsExpected:     255,
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{Type: SegmentationUPIDTypeEIDR, Value: "10.5239/F9B9-7B34-0000-0000-0000"},
							{Type: SegmentationUPIDTypeADS, Value: "type=LA&dur=60000&tierü0"},
							{Type: uint32(199)},
							{Type: uint32(0)},
							{Type: uint32(0)},
							{Type: uint32(0)},
							{Type: uint32(255)},
						},
					},
				},
				PTSAdjustment: uint64(5940),
				Tier:          4095,
				SAPType:       3,
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			sis, err := DecodeBase64(c.binary)
			if err != nil {
				t.Fatal(err)
			}
			// legacy 35's produce an "updated" binary so will not match
			if !c.legacy {
				if c.binary != sis.Base64() {
					t.Errorf("re-encode to binary: want %s, got %s", c.binary, sis.Base64())
				}
			}
		})
	}
}

func TestBadBase64(t *testing.T) {
	var tests = []struct {
		name string
		in   string
		err  error
	}{
		{"empty", "", ErrBufferOverflow},
		{"invalid base64", "/DBaf%^", ErrUnsupportedEncoding},
		{"Invalid CRC_32", "/DA4AAAAAAAAAP/wFAUABDEAf+//mWEhzP4Azf5gAQAAAAATAhFDVUVJAAAAAX+/AQIwNAEAAKeYO3Q=", ErrCRC32Invalid},
		{
			"SpliceInsert time with invalid component count",
			"FkC1lwP3uTQD0VvxHwVBEH89G6B7VjzaZ9eNuyUF9q8pYAIXsRM9ZpDCczBeDbytQhXkssQstGJVGcvjZ3tiIMULiA4BpRHlzLGFa0q6aVMtzk8ZRUeLcxtKibgVOKBBnkCbOQyhSflFiDkrAAIp+Fk+VRsByTSkPN3RvyK+lWcjHElhwa9hNFcAy4dm3DdeRXnrD3I2mISNc7DkgS0ReotPyp94FV77xMHT4D7SYL48XU20UM4bgg==",
			ErrBufferOverflow,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeBase64(tt.in)
			if !errors.Is(err, tt.err) {
				t.Errorf("DecodeBase64: want error %v, got %v", tt.err, err)
			}
		})
	}
}

func TestDecodeHex(t *testing.T) {
	// when adding tests that contain multiple splice descriptors, care must be
	// taken to ensure they are in the order specified in the custom UnmarshalXML
	// implementation, otherwise misleading error may occur
	cases := map[string]struct {
		hex      string
		err      error
		expected SpliceInfoSection
	}{
		"Sample 14.1 time_signal - Placement Opportunity Start": {
			hex: "0xFC3034000000000000FFFFF00506FE72BD0050001E021C435545494800008E7FCF0001A599B00808000000002CA0A18A3402009AC9D17E",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{
						PTSTime: uint64ptr(0x072bd0050),
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							NoRegionalBlackoutFlag: true,
							ArchiveAllowedFlag:     true,
							DeviceRestrictions:     DeviceRestrictionsNone,
						},
						SegmentationEventID:  uint32(0x4800008e),
						SegmentationTypeID:   SegmentationTypeProviderPOStart,
						SegmentationDuration: uint64ptr(0x0001a599b0),
						SegmentationUPIDs: []SegmentationUPID{
							NewSegmentationUPID(SegmentationUPIDTypeTI, toBytes(0x000000002ca0a18a)),
						},
						SegmentNum: 2,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.2 splice_insert (no prefix)": {
			hex: "FC302F000000000000FFFFF014054800008F7FEFFE7369C02EFE0052CCF500000000000A0008435545490000013562DBA30A",
			expected: SpliceInfoSection{
				EncryptedPacket: EncryptedPacket{EncryptionAlgorithm: EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &SpliceInsert{
					BreakDuration: &BreakDuration{
						AutoReturn: true,
						Duration:   uint64(0x00052ccf5),
					},
					SpliceEventID:         uint32(0x4800008f),
					OutOfNetworkIndicator: true,
					Program: &SpliceInsertProgram{
						SpliceTime: SpliceTime{
							PTSTime: uint64ptr(0x07369c02e),
						},
					},
				},
				SpliceDescriptors: []SpliceDescriptor{
					&AvailDescriptor{
						ProviderAvailID: 0x00000135,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			_, err := DecodeHex(c.hex)
			if !errors.Is(c.err, err) {
				t.Fatalf("want error %v, got %v", c.err, err)
			}
		})
	}
}

func TestEncodeWithAlignmentStuffing(t *testing.T) {
	cases := map[string]struct {
		name   string
		binary string
	}{
		"SpliceInsert Program Out Point with 3 bytes alignment stuffing": {
			binary: "/DA0AABS2+YAAACgFAUALJGCf+/+MSwPcX4AUmXAAAAAAAAMAQpDVUVJRp8xMjEq3pnIPCi6lw==",
		},
		"SpliceInsert Program In Point with 3 bytes alignment stuffing": {
			binary: "/DAvAABS2+YAAACgDwUALJGEf0/+MX7z3AAAAAAADAEKQ1VFSQCfMTIxI6SMuQkzWQI=",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sis, err := DecodeBase64(c.binary)
			if err != nil {
				t.Fatal(err)
			}
			if c.binary != sis.Base64() {
				t.Errorf("base64 re-encode: want %s, got %s", c.binary, sis.Base64())
			}
		})
	}
}
*/

func TestTicksToDuration(t *testing.T) {
	// test a wide range of tick values
	min := 29 * TicksPerSecond
	max := 61 * TicksPerSecond
	for i := min; i < max; i++ {
		d := TicksToDuration(uint64(i))
		ticks := DurationToTicks(d)
		if i != int(ticks) {
			t.Errorf("DurationToTicks(%s) = %d, want %d", d, ticks, i)
		}
	}
}

// helper func to make test life a bit easier

func toBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func uint32ptr(i uint32) *uint32 {
	return &i
}

func uint64ptr(i uint64) *uint64 {
	return &i
}
