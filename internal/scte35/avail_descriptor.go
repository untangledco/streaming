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

	"github.com/bamiaux/iobit"
)

const (
	// AvailDescriptorTag is the splice_descriptor_tag for an avail_descriptor
	AvailDescriptorTag = 0x00
)

// AvailDescriptor is an implementation of a splice_descriptor. It provides an
// optional extension to the splice_insert() command that allows an
// authorization identifier to be sent for an avail. Multiple copies of this
// descriptor may be included by using the loop mechanism provided. This
// identifier is intended to replicate the functionality of the cue tone system
// used in analog systems for ad insertion. This descriptor is intended only
// for use with a splice_insert() command, within a splice_info_section.
type AvailDescriptor struct {
	XMLName         xml.Name `xml:"http://www.scte.org/schemas/35 AvailDescriptor" json:"-"`
	JSONType        uint32   `xml:"-" json:"type"`
	ProviderAvailID uint32   `xml:"providerAvailId,attr" json:"providerAvailId"`
}

// Tag returns the splice_descriptor_tag.
func (sd *AvailDescriptor) Tag() uint32 {
	// ensure JSONType is set
	sd.JSONType = AvailDescriptorTag
	return AvailDescriptorTag
}

// writeTo the given table.
func (sd *AvailDescriptor) writeTo(t *table) {
	t.row(0, "avail_descriptor() {", nil)
	t.row(1, "splice_descriptor_tag", fmt.Sprintf("%#02x", AvailDescriptorTag))
	t.row(1, "descriptor_length", sd.length())
	t.row(1, "identifier", fmt.Sprintf("%#08x, (%s)", CUEIdentifier, CUEIASCII))
	t.row(1, "provider_avail_id", sd.ProviderAvailID)
	t.row(0, "}", nil)
}

// decode updates this splice_descriptor from binary.
func (sd *AvailDescriptor) decode(b []byte) error {
	r := iobit.NewReader(b)
	r.Skip(8)  // splice_descriptor_tag
	r.Skip(8)  // descriptor_length
	r.Skip(32) // identifier
	sd.ProviderAvailID = r.Uint32(32)

	if err := readerError(r); err != nil {
		return fmt.Errorf("avail_descriptor: %w", err)
	}
	return nil
}

// encode this splice_descriptor to binary.
func (sd *AvailDescriptor) encode() ([]byte, error) {
	length := sd.length()
	// add 2 bytes to contain splice_descriptor_tag & descriptor_length
	buf := make([]byte, length+2)
	iow := iobit.NewWriter(buf)

	iow.PutUint32(8, AvailDescriptorTag)  // splice_descriptor_tag
	iow.PutUint32(8, uint32(length))      // descriptor_length
	iow.PutUint32(32, CUEIdentifier)      // identifier
	iow.PutUint32(32, sd.ProviderAvailID) // provider_avail_id

	err := iow.Flush()
	return buf, err
}

// descriptorLength returns the descriptor_length
func (sd *AvailDescriptor) length() int {
	length := 32 // identifier
	length += 32 // provider_avail_id
	return length / 8
}
