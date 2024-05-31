package scte35

type sample struct {
	name    string
	encoded string
	want    Splice
}

var samples = []sample{
	{
		name:    "14.1. time_signal",
		encoded: "/DA0AAAAAAAA///wBQb+cr0AUAAeAhxDVUVJSAAAjn/PAAGlmbAICAAAAAAsoKGKNAIAmsnRfg==",
		want: Splice{
			SAPType: SAPNone,
			Tier:    0x0fff,
			CWIndex: 0xff,
			Command: &Command{
				Type:       TimeSignal,
				TimeSignal: newuint64(0x072bd0050),
			},
			Descriptors: []SpliceDescriptor{
				SegmentationDescriptor{
					EventID:      0x4800008e,
					idCompliance: true,
					Restrictions: NoRegionalBlackout | ArchiveAllowed | DeviceRestrictionsNone,
					Duration:     newuint64(0x0001a599b0),
					UPID: UPID{
						Type:  UPIDTI,
						Value: []byte{0x00, 0x00, 0x00, 0x00, 0x2c, 0xa0, 0xa1, 0x8a},
					},
					Type:     ProviderPlacementOppStart,
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
		want: Splice{
			SAPType: SAPNone,
			CWIndex: 0xff,
			Tier:    0x0fff,
			Command: &Command{
				Type: SpliceInsert,
				Insert: &Insert{
					ID:           0x4800008f,
					idCompliance: true,
					OutOfNetwork: true,
					SpliceTime:   newuint64(0x07369c02e),
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
	{
		name:    "14.3. time_signal",
		encoded: "/DAvAAAAAAAA///wBQb+dGKQoAAZAhdDVUVJSAAAjn+fCAgAAAAALKChijUCAKnMZ1g=",
		want: Splice{
			SAPType: SAPNone,
			CWIndex: 0xff,
			Tier:    0x0fff,
			Command: &Command{
				Type:       TimeSignal,
				TimeSignal: newuint64(0x0746290a0),
			},
			Descriptors: []SpliceDescriptor{
				SegmentationDescriptor{
					EventID:      0x4800008e,
					idCompliance: true,
					Restrictions: WebDeliveryAllowed | NoRegionalBlackout | ArchiveAllowed | DeviceRestrictionsNone,
					UPID: UPID{
						Type:  UPIDTI,
						Value: []byte{0x00, 0x00, 0x00, 0x00, 0x2c, 0xa0, 0xa1, 0x8a},
					},
					Type:   ProviderPlacementOppEnd,
					Number: 2,
				},
			},
			CRC32: 0xa9cc6758,
		},
	},
	{
		name:    "14.4. time_signal",
		encoded: "/DBIAAAAAAAA///wBQb+ek2ItgAyAhdDVUVJSAAAGH+fCAgAAAAALMvDRBEAAAIXQ1VFSUgAABl/nwgIAAAAACyk26AQAACZcuND",
		want: Splice{
			SAPType: SAPNone,
			CWIndex: 0xff,
			Tier:    0x0fff,
			Command: &Command{
				Type:       TimeSignal,
				TimeSignal: newuint64(0x07a4d88b6),
			},
			Descriptors: []SpliceDescriptor{
				SegmentationDescriptor{
					EventID:      0x48000018,
					Restrictions: WebDeliveryAllowed | NoRegionalBlackout | ArchiveAllowed | DeviceRestrictionsNone,
					idCompliance: true,
					UPID: UPID{
						Type:  UPIDTI,
						Value: []byte{0, 0, 0, 0, 0x2c, 0xcb, 0xc3, 0x44},
					},
					Type: ProgramEnd,
				},
				SegmentationDescriptor{
					EventID:      0x48000019,
					idCompliance: true,
					Restrictions: WebDeliveryAllowed | NoRegionalBlackout | ArchiveAllowed | DeviceRestrictionsNone,
					UPID: UPID{
						Type:  UPIDTI,
						Value: []byte{0, 0, 0, 0, 0x2c, 0xa4, 0xdb, 0xa0},
					},
					Type: ProgramStart,
				},
			},
			CRC32: 0x9972e343,
		},
	},
	{
		// from "The Essential Guide to SCTE-35" by Bitmovin (https://bitmovin.com/scte-35-guide)
		name:    "dtmf",
		encoded: "/DBcAAAAAAAAAP/wBQb//ciI8QBGAh1DVUVJXQk9EX+fAQ5FUDAxODAzODQwMDY2NiEEZAIZQ1VFSV0JPRF/3wABLit7AQVDMTQ2NDABAQEKQ1VFSQCAMTUwKnPhdcU=",
		want: Splice{
			SAPType: SAPNone,
			Tier:    0x0fff,
			Command: &Command{
				Type:       TimeSignal,
				TimeSignal: newuint64(8552745201),
			},
			Descriptors: []SpliceDescriptor{
				SegmentationDescriptor{
					EventID:      1560886545,
					idCompliance: true,
					Restrictions: WebDeliveryAllowed | NoRegionalBlackout | DeviceRestrictionsNone,
					UPID: UPID{
						Type:  UPIDType(1),
						Value: []byte{69, 80, 48, 49, 56, 48, 51, 56, 52, 48, 48, 54, 54, 54},
					},
					Type:     ChapterEnd,
					Number:   4,
					Expected: 100,
				},
				SegmentationDescriptor{
					EventID:      1560886545,
					idCompliance: true,
					Restrictions: WebDeliveryAllowed | NoRegionalBlackout | DeviceRestrictionsNone,
					Duration:     newuint64(19803003),
					UPID: UPID{
						Type:  UPIDType(1),
						Value: []byte{67, 49, 52, 54, 52},
					},
					Type:     ProviderAdStart,
					Number:   1,
					Expected: 1,
				},
				DTMFDescriptor{Chars: []byte("150*")},
			},
			CRC32: 1944155589,
		},
	},
}
