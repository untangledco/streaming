package scte35

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpliceInfoSection_UnmarshalXML(t *testing.T) {
	cases := map[string]struct {
		xml      string
		expected *SpliceInfoSection
	}{
		"SAPType Missing": {
			xml: `
				<SpliceInfoSection xmlns="http://www.scte.org/schemas/35" tier="4095">
					<EncryptedPacket xmlns="http://www.scte.org/schemas/35" cwIndex="255"></EncryptedPacket>
					<TimeSignal xmlns="http://www.scte.org/schemas/35">
						<SpliceTime xmlns="http://www.scte.org/schemas/35" ptsTime="1924989008"></SpliceTime>
					</TimeSignal>
					<SegmentationDescriptor xmlns="http://www.scte.org/schemas/35" segmentationEventId="1207959694" segmentationDuration="27630000" segmentationTypeId="52" segmentNum="2">
						<DeliveryRestrictions xmlns="http://www.scte.org/schemas/35" archiveAllowedFlag="true" webDeliveryAllowedFlag="false" noRegionalBlackoutFlag="true" deviceRestrictions="3"></DeliveryRestrictions>
						<SegmentationUpid xmlns="http://www.scte.org/schemas/35" segmentationUpidType="8">748724618</SegmentationUpid>
					</SegmentationDescriptor>
				</SpliceInfoSection>`,
			expected: &SpliceInfoSection{
				Tier:            uint32(4095),
				SAPType:         SAPTypeNotSpecified,
				EncryptedPacket: EncryptedPacket{CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{PTSTime: uint64ptr(1924989008)},
				},
				SpliceDescriptors: SpliceDescriptors{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:  SegmentationUPIDTypeTI,
								Value: "748724618",
							},
						},
						SegmentationEventID:  uint32(1207959694),
						SegmentationDuration: uint64ptr(27630000),
						SegmentationTypeID:   SegmentationTypeProviderPOStart,
						SegmentNum:           2,
					},
				},
			},
		},
		"SAPType Specified": {
			xml: `
				<SpliceInfoSection xmlns="http://www.scte.org/schemas/35" tier="4095" sapType="0">
					<EncryptedPacket xmlns="http://www.scte.org/schemas/35" cwIndex="255"></EncryptedPacket>
					<TimeSignal xmlns="http://www.scte.org/schemas/35">
						<SpliceTime xmlns="http://www.scte.org/schemas/35" ptsTime="1924989008"></SpliceTime>
					</TimeSignal>
					<SegmentationDescriptor xmlns="http://www.scte.org/schemas/35" segmentationEventId="1207959694" segmentationDuration="27630000" segmentationTypeId="52" segmentNum="2">
						<DeliveryRestrictions xmlns="http://www.scte.org/schemas/35" archiveAllowedFlag="true" webDeliveryAllowedFlag="false" noRegionalBlackoutFlag="true" deviceRestrictions="3"></DeliveryRestrictions>
						<SegmentationUpid xmlns="http://www.scte.org/schemas/35" segmentationUpidType="8">748724618</SegmentationUpid>
					</SegmentationDescriptor>
				</SpliceInfoSection>`,
			expected: &SpliceInfoSection{
				Tier:            uint32(4095),
				SAPType:         SAPType1,
				EncryptedPacket: EncryptedPacket{CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{PTSTime: uint64ptr(1924989008)},
				},
				SpliceDescriptors: SpliceDescriptors{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:  SegmentationUPIDTypeTI,
								Value: "748724618",
							},
						},
						SegmentationEventID:  uint32(1207959694),
						SegmentationDuration: uint64ptr(27630000),
						SegmentationTypeID:   SegmentationTypeProviderPOStart,
						SegmentNum:           2,
					},
				},
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var sis SpliceInfoSection
			require.NoError(t, xml.Unmarshal([]byte(c.xml), &sis))
			require.Equal(t, toXML(c.expected), toXML(&sis))
		})
	}
}

func TestSpliceInfoSection_UnmarshalJSON(t *testing.T) {
	cases := map[string]struct {
		json     string
		expected *SpliceInfoSection
	}{
		"SAPType Missing": {
			json: `{
				"encryptedPacket": {
					"cwIndex": 255
				},
				"spliceCommand": {
					"type": 6,
					"spliceTime": {
						"ptsTime": 1924989008
					}
				},
				"spliceDescriptors": [
					{
						"type": 2,
						"deliveryRestrictions": {
							"archiveAllowedFlag": true,
							"webDeliveryAllowedFlag": false,
							"noRegionalBlackoutFlag": true,
							"deviceRestrictions": 3
						},
						"segmentationUpids": [
							{
								"segmentationUpidType": 8,
								"format": "text",
								"value": "748724618"
							}
						],
						"segmentationEventId": 1207959694,
						"segmentationDuration": 27630000,
						"segmentationTypeId": 52,
						"segmentNum": 2
					}
				],
				"tier": 4095
			}`,
			expected: &SpliceInfoSection{
				Tier:            uint32(4095),
				SAPType:         SAPTypeNotSpecified,
				EncryptedPacket: EncryptedPacket{CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{PTSTime: uint64ptr(1924989008)},
				},
				SpliceDescriptors: SpliceDescriptors{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:  SegmentationUPIDTypeTI,
								Value: "748724618",
							},
						},
						SegmentationEventID:  uint32(1207959694),
						SegmentationDuration: uint64ptr(27630000),
						SegmentationTypeID:   SegmentationTypeProviderPOStart,
						SegmentNum:           2,
					},
				},
			},
		},
		"SAPType Specified": {
			json: `{
				"encryptedPacket": {
					"cwIndex": 255
				},
				"sapType": 0,
				"spliceCommand": {
					"type": 6,
					"spliceTime": {
						"ptsTime": 1924989008
					}
				},
				"spliceDescriptors": [
					{
						"type": 2,
						"deliveryRestrictions": {
							"archiveAllowedFlag": true,
							"webDeliveryAllowedFlag": false,
							"noRegionalBlackoutFlag": true,
							"deviceRestrictions": 3
						},
						"segmentationUpids": [
							{
								"segmentationUpidType": 8,
								"format": "text",
								"value": "748724618"
							}
						],
						"segmentationEventId": 1207959694,
						"segmentationDuration": 27630000,
						"segmentationTypeId": 52,
						"segmentNum": 2
					}
				],
				"tier": 4095
			}`,
			expected: &SpliceInfoSection{
				Tier:            uint32(4095),
				SAPType:         SAPType1,
				EncryptedPacket: EncryptedPacket{CWIndex: 255},
				SpliceCommand: &TimeSignal{
					SpliceTime: SpliceTime{PTSTime: uint64ptr(1924989008)},
				},
				SpliceDescriptors: SpliceDescriptors{
					&SegmentationDescriptor{
						DeliveryRestrictions: &DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []SegmentationUPID{
							{
								Type:  SegmentationUPIDTypeTI,
								Value: "748724618",
							},
						},
						SegmentationEventID:  uint32(1207959694),
						SegmentationDuration: uint64ptr(27630000),
						SegmentationTypeID:   SegmentationTypeProviderPOStart,
						SegmentNum:           2,
					},
				},
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var sis SpliceInfoSection
			require.NoError(t, json.Unmarshal([]byte(c.json), &sis))
			require.Equal(t, toXML(c.expected), toXML(&sis))
		})
	}
}
