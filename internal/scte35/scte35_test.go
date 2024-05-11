package scte35

type sample struct {
	name    string
	encoded string
	want    SpliceInfo
}

var samples = []sample{
	{
		name:    "14.1. time_signal",
		encoded: "/DA0AAAAAAAA///wBQb+cr0AUAAeAhxDVUVJSAAAjn/PAAGlmbAICAAAAAAsoKGKNAIAmsnRfg==",
		want: SpliceInfo{
			SAPType: SAPNone,
			Tier:    0x0fff,
			Command: &Command{
				Type:       TimeSignal,
				TimeSignal: newuint64(0x072bd0050),
			},
			Descriptors: []SpliceDescriptor{
				SegmentationDescriptor{
					EventID:           0x4800008e,
					EventIDCompliance: true,
					Restrictions:      NoRegionalBlackout | ArchiveAllowed | DeviceRestrictionsNone,
					Duration:          newuint64(0x0001a599b0),
					UPID: UPID{
						Type:  UPIDTI,
						Value: []byte{0x00, 0x00, 0x00, 0x00, 0x2c, 0xa0, 0xa1, 0x8a},
					},
					Type:     0x34,
					Number:   2,
					Expected: 0,
				},
			},
			CRC32: 0x9ac9d17e,
		},
	},
	{
		name:    "14.2. splice_insert",
		encoded: "/DAvAAAAAAAA///wFAVIAACPf+/+c2nALv4AUsz1AAAAAAAKAAhDVUVJAAABNWLbowo=",
		want: SpliceInfo{
			SAPType: SAPNone,
			Tier:    0x0fff,
			Command: &Command{
				Type: SpliceInsert,
				Insert: &Insert{
					ID:                0x4800008f,
					EventIDCompliance: true,
					OutOfNetwork:      true,
					SpliceTime:        newuint64(0x07369c02e),
					Duration: &BreakDuration{
						AutoReturn: true,
						Duration:   0x00052ccf5,
					},
				},
			},
			Descriptors: []SpliceDescriptor{
				AvailDescriptor(0x00000135),
			},
			CRC32: 0x62dba30a,
		},
	},
}

func newuint64(i uint64) *uint64 { p := new(uint64); p = &i; return p }
