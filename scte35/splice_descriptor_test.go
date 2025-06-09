package scte35

import (
	"reflect"
	"testing"
)

func TestAudioDescriptorRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		desc AudioDescriptor
	}{
		{
			name: "empty",
			desc: AudioDescriptor{},
		},
		{
			name: "single channel",
			desc: AudioDescriptor{
				{
					ComponentTag:  0x12,
					Language:      [3]byte{'e', 'n', 'g'},
					BitstreamMode: 0x05,
					Count:         6,
					FullService:   true,
				},
			},
		},
		{
			name: "multiple channels",
			desc: AudioDescriptor{
				{
					ComponentTag:  0x12,
					Language:      [3]byte{'e', 'n', 'g'},
					BitstreamMode: 0x05,
					Count:         6,
					FullService:   true,
				},
				{
					ComponentTag:  0x34,
					Language:      [3]byte{'s', 'p', 'a'},
					BitstreamMode: 0x03,
					Count:         2,
					FullService:   false,
				},
				{
					ComponentTag:  0x56,
					Language:      [3]byte{'f', 'r', 'a'},
					BitstreamMode: 0x07,
					Count:         8,
					FullService:   true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic properties
			if got := tt.desc.Tag(); got != TagAudio {
				t.Errorf("Tag() = %v, want %v", got, TagAudio)
			}
			if got := tt.desc.ID(); got != descriptorIDCUEI {
				t.Errorf("ID() = %v, want %v", got, descriptorIDCUEI)
			}

			// Test round-trip marshalling/unmarshalling
			data := tt.desc.Data()
			unmarshaled := unmarshalAudioDescriptor(data)

			if !reflect.DeepEqual(tt.desc, unmarshaled) {
				t.Errorf("Round-trip failed:\noriginal:    %+v\nunmarshaled: %+v", tt.desc, unmarshaled)
			}
		})
	}
}

func TestAudioDescriptorUnmarshalSpliceDescriptor(t *testing.T) {
	// Test that AudioDescriptor can be unmarshalled via unmarshalSpliceDescriptor
	desc := AudioDescriptor{
		{
			ComponentTag:  0x12,
			Language:      [3]byte{'e', 'n', 'g'},
			BitstreamMode: 0x05,
			Count:         6,
			FullService:   true,
		},
	}

	// Encode as a full splice descriptor
	encoded := encodeSpliceDescriptor(desc)

	// Unmarshal via the main function
	unmarshaled, err := unmarshalSpliceDescriptor(encoded)
	if err != nil {
		t.Fatalf("unmarshalSpliceDescriptor failed: %v", err)
	}

	// Check that we got back an AudioDescriptor
	audioDesc, ok := unmarshaled.(AudioDescriptor)
	if !ok {
		t.Fatalf("Expected AudioDescriptor, got %T", unmarshaled)
	}

	if !reflect.DeepEqual(desc, audioDesc) {
		t.Errorf("Round-trip via unmarshalSpliceDescriptor failed:\noriginal:    %+v\nunmarshaled: %+v", desc, audioDesc)
	}
}

func TestAudioDescriptorDataFormat(t *testing.T) {
	// Test specific encoding format
	desc := AudioDescriptor{
		{
			ComponentTag:  0x12,
			Language:      [3]byte{'e', 'n', 'g'},
			BitstreamMode: 0x05, // 3 bits: 101
			Count:         6,    // 4 bits: 0110
			FullService:   true, // 1 bit: 1
		},
	}

	data := desc.Data()
	expected := []byte{
		0x10,          // count=1 in upper 4 bits: 00010000
		0x12,          // ComponentTag
		'e', 'n', 'g', // Language
		0xAD, // BitstreamMode(101) + Count(0110) + FullService(1) = 10101101 = 0xAD
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Data() encoding mismatch:\ngot:      %x\nexpected: %x", data, expected)
	}

	// Test unmarshalling the expected format
	unmarshaled := unmarshalAudioDescriptor(expected)
	if !reflect.DeepEqual(desc, unmarshaled) {
		t.Errorf("Unmarshalling expected format failed:\noriginal:    %+v\nunmarshaled: %+v", desc, unmarshaled)
	}
}
