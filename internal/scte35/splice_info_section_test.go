package scte35

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
