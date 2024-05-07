package scte35

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestPackTier(t *testing.T) {
	want := maxTier - 1
	fmt.Printf("%012b\n", want)

	packed := packTier(want)
	got := uint16(packed[1]) | uint16(packed[0])<<8
	if got != want {
		t.Errorf("want packed tier %d, got %d", want, got)
	}
}

func TestReadSpliceInfo(t *testing.T) {
	var tsig uint64 = 0x072bd0050
	var segdur uint64 = 0x0001a599b0
	sdesc := SegmentationDescriptor{
		EventID:      0x4800008e,
		Restrictions: NoRegionalBlackout | ArchiveAllowed | DeviceRestrictGroup2,
		Duration:     &segdur,
		UPID: UPID{
			UPIDTypeTI,
			[]byte{0x8a, 0xa1, 0xa0, 0x2c, 0x00, 0x00, 0x00, 0x00},
		},
		Type:   0x34,
		Number: 2,
	}
	sis := SpliceInfo{
		Tier: 0xfff,
		Command: &Command{
			Type:       TimeSignal,
			TimeSignal: &tsig,
		},
		Descriptors: []SpliceDescriptor{
			{
				Tag:  TagSegmentation,
				ID:   binary.LittleEndian.Uint32([]byte(DescriptorIDCUEI)),
				Data: encodeSegDescriptor(&sdesc),
			},
		},
		CRC32: 0x9ac9d17e,
	}

	want := "/DAvAAAAAAAA///wFAVIAACPf+/+c2nALv4AUsz1AAAAAAAKAAhDVUVJAAABNWLbowo="

	bgot, err := encodeSpliceInfo(&sis)
	if err != nil {
		t.Fatalf("encode splice info: %v", err)
	}
	got := base64.StdEncoding.EncodeToString(bgot)
	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}
