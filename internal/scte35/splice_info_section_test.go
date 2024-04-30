package scte35

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path"
	"strings"
	"testing"
)

type sistest struct {
	name string
	sis  SpliceInfoSection
}

var tsis SpliceInfoSection = SpliceInfoSection{
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
}

func TestSpliceInfoSection(t *testing.T) {
	missingSAP := tsis
	withSAP := tsis
	withSAP.SAPType = SAPType1

	var tests = []sistest{
		{"testdata/saptype_missing.xml", missingSAP},
		{"testdata/saptype_missing.json", missingSAP},
		{"testdata/saptype_specified.xml", withSAP},
		{"testdata/saptype_specified.json", withSAP},
	}

	for _, tt := range tests {
		tname := path.Base(tt.name)
		t.Run(tname, func(t *testing.T) {
			b, err := os.ReadFile(tt.name)
			if err != nil {
				t.Fatal(err)
			}
			var got SpliceInfoSection
			if strings.HasSuffix(tt.name, "json") {
				if err := json.Unmarshal(b, &got); err != nil {
					t.Fatal(err)
				}
				if toJSON(&tt.sis) != toJSON(&got) {
					t.Error("remarshalled json different from source")
					t.Logf("want: %s", toJSON(&tt.sis))
					t.Logf("got: %s", toJSON(&got))
				}
			} else {
				if err := xml.Unmarshal(b, &got); err != nil {
					t.Fatal(err)
				}
				if toXML(&tt.sis) != toXML(&got) {
					t.Error("remarshalled xml different from source")
					t.Logf("want: %s", toXML(&tt.sis))
					t.Logf("got: %s", toXML(&got))
				}
			}
		})
	}
}
