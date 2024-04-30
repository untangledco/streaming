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
	// DTMFDescriptorTag is the splice_descriptor_tag for a dtmf_descriptor
	DTMFDescriptorTag = 0x01
)

// DTMFDescriptor is an implementation of a splice_descriptor. It provides an
// optional extension to the splice_insert() command that allows a receiver
// device to generate a legacy analog DTMF sequence based on a
// splice_info_section being received.
type DTMFDescriptor struct {
	XMLName   xml.Name `xml:"http://www.scte.org/schemas/35 DTMFDescriptor" json:"-"`
	JSONType  uint32   `xml:"-" json:"type"`
	Preroll   uint32   `xml:"preroll,attr" json:"preroll"`
	DTMFChars string   `xml:"chars,attr" json:"chars"`
}

// writeTo the given table.
func (sd *DTMFDescriptor) writeTo(t *table) {
	t.row(0, "dtmf_descriptor() {", nil)
	t.row(1, "splice_descriptor_tag", fmt.Sprintf("%#02x", DTMFDescriptorTag))
	t.row(1, "descriptor_length", sd.length())
	t.row(1, "identifier", fmt.Sprintf("%#08x (%s)", CUEIdentifier, CUEIASCII))
	t.row(1, "preroll", float32(sd.Preroll/10))
	t.row(1, "dtmf_count", len(sd.DTMFChars))
	t.row(1, "dtmf_chars", sd.DTMFChars)
	t.row(0, "}", nil)
}

// Tag returns the splice_descriptor_tag.
func (sd *DTMFDescriptor) Tag() uint32 {
	// ensure JSONType is set
	sd.JSONType = DTMFDescriptorTag
	return DTMFDescriptorTag
}

// decode updates this splice_descriptor from binary.
func (sd *DTMFDescriptor) decode(b []byte) error {
	r := iobit.NewReader(b)
	r.Skip(8)  // splice_descriptor_tag
	r.Skip(8)  // descriptor_length
	r.Skip(32) // identifier
	sd.Preroll = r.Uint32(8)
	dtmfCount := int(r.Uint32(3))
	r.Skip(5) // reserved
	sd.DTMFChars = r.String(dtmfCount)

	if err := readerError(r); err != nil {
		return fmt.Errorf("dtmf_descriptor: %w", err)
	}
	return readerError(r)
}

// encode this splice_descriptor to binary.
func (sd *DTMFDescriptor) encode() ([]byte, error) {
	length := sd.length()

	// add 2 bytes to contain splice_descriptor_tag & descriptor_length
	buf := make([]byte, length+2)
	iow := iobit.NewWriter(buf)
	iow.PutUint32(8, DTMFDescriptorTag)         // splice_descriptor_tag
	iow.PutUint32(8, uint32(length))            // descriptor_length
	iow.PutUint32(32, CUEIdentifier)            // identifier
	iow.PutUint32(8, sd.Preroll)                // preroll
	iow.PutUint32(3, uint32(len(sd.DTMFChars))) // dtmf_count
	iow.PutUint32(5, Reserved)                  // reserved
	_, err := iow.Write([]byte(sd.DTMFChars))   // dtmf_chars
	if err != nil {
		return buf, err
	}
	err = iow.Flush()
	return buf, err
}

// descriptorLength returns the descriptor_length.
func (sd *DTMFDescriptor) length() int {
	length := 32                    // identifier
	length += 8                     // preroll
	length += 3                     // dtmf_count
	length += 5                     // reserved
	length += len(sd.DTMFChars) * 8 // dtmf_char
	return length / 8
}
