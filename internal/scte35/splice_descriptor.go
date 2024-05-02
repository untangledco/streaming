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

import "github.com/bamiaux/iobit"

const (
	// CUEIdentifier is 32-bit number used to identify the owner of the
	// descriptor. The identifier shall have a value of 0x43554549 (ASCII “CUEI”).
	CUEIdentifier = 0x43554549
	// CUEIASCII is the CUEIIdentifier ASCII value
	CUEIASCII = "CUEI"
)

// NewSpliceDescriptor returns the appropriate splice_descriptor for the given
// identifier and tag
func NewSpliceDescriptor(identifier uint32, tag uint32) SpliceDescriptor {
	if identifier == CUEIdentifier {
		switch tag {
		case AvailDescriptorTag:
			return &AvailDescriptor{}
		case DTMFDescriptorTag:
			return &DTMFDescriptor{}
		case SegmentationDescriptorTag:
			return &SegmentationDescriptor{}
		case TimeDescriptorTag:
			return &TimeDescriptor{}
		case AudioDescriptorTag:
			return &AudioDescriptor{}
		}
	}
	// as a last resort, fall back to private_descriptor. This is not strictly
	// compliant but allows us to deal with a wider array of quirky signals.
	return &PrivateDescriptor{Identifier: identifier}

}

// SpliceDescriptor is a prototype for adding new fields to the
// splice_info_section. All descriptors included use the same syntax for the
// first six bytes. In order to allow private information to be added we have
// included the ‘identifier’ code. This removes the need for a registration
// descriptor in the descriptor loop.
//
// Any receiving equipment should skip any descriptors with unknown identifiers
// or unknown descriptor tags. For descriptors with known identifiers, the
// receiving equipment should skip descriptors with an unknown
// splice_descriptor_tag.
type SpliceDescriptor interface {
	Tag() uint32
	decode(b []byte) error
	encode() ([]byte, error)
	length() int // named to differentiate from splice_command
}

// SpliceDescriptors is a slice of SpliceDescriptor.
type SpliceDescriptors []SpliceDescriptor

// decodeSpliceDescriptors returns a slice of SpliceDescriptors from decoding
// the supplied byte array.
func decodeSpliceDescriptors(b []byte) ([]SpliceDescriptor, error) {
	r := iobit.NewReader(b)

	var sds []SpliceDescriptor
	for r.LeftBits() > 0 {
		// Peek to get splice_descriptor_tag, descriptor_length, and
		// identifier
		sdr := r.Peek()
		spliceDescriptorTag := sdr.Uint32(8)
		descriptorLength := int(sdr.Uint32(8))
		identifier := sdr.Uint32(32)

		// Decode the full splice_descriptor (including splice_descriptor_tag
		// and descriptor_length).
		sd := NewSpliceDescriptor(identifier, spliceDescriptorTag)
		err := sd.decode(r.Bytes(descriptorLength + 2))
		if err != nil {
			return sds, err
		}
		sds = append(sds, sd)
	}

	return sds, nil
}
