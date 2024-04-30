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
	"encoding/binary"
	"encoding/xml"
	"fmt"

	"github.com/bamiaux/iobit"
)

// PrivateDescriptor encapsulates the contents of non-CUEI descriptors
type PrivateDescriptor struct {
	XMLName      xml.Name `xml:"http://www.scte.org/schemas/35 PrivateDescriptor" json:"-"`
	JSONType     uint32   `xml:"-" json:"type"`
	PrivateTag   uint32   `xml:"tag,attr" json:"tag"`
	Identifier   uint32   `xml:"identifier,attr" json:"identifier"`
	PrivateBytes Bytes    `xml:",chardata" json:"privateBytes"`
}

// IdentifierString returns the identifier as a string
func (sd *PrivateDescriptor) IdentifierString() string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, sd.Identifier)
	return string(b)
}

// Tag returns the splice_descriptor_tag.
func (sd *PrivateDescriptor) Tag() uint32 {
	// ensure JSONType is set
	sd.JSONType = sd.PrivateTag
	return sd.PrivateTag
}

// decode updates this splice_descriptor from binary.
func (sd *PrivateDescriptor) decode(b []byte) error {
	r := iobit.NewReader(b)

	sd.PrivateTag = r.Uint32(8)
	r.Skip(8) // descriptor_length
	sd.Identifier = r.Uint32(32)
	sd.PrivateBytes = r.LeftBytes()
	// LeftBytes doesnt advance position
	r.Skip(uint(len(sd.PrivateBytes) * 8))

	if err := readerError(r); err != nil {
		return fmt.Errorf("private_descriptor: %w", err)
	}
	return nil
}

// encode this splice_descriptor to binary.
func (sd *PrivateDescriptor) encode() ([]byte, error) {
	length := sd.length()

	// add 2 bytes to contain splice_descriptor_tag & descriptor_length
	buf := make([]byte, length+2)
	iow := iobit.NewWriter(buf)
	iow.PutUint32(8, sd.PrivateTag)
	iow.PutUint32(8, uint32(length))
	iow.PutUint32(32, sd.Identifier)
	_, err := iow.Write(sd.PrivateBytes)
	if err != nil {
		return buf, err
	}
	err = iow.Flush()

	return buf, err
}

// descriptorLength returns the descriptor_length
func (sd *PrivateDescriptor) length() int {
	length := 32                       // identifier
	length += len(sd.PrivateBytes) * 8 // private_bytes
	return length / 8
}

// table returns the tabular description of this PrivateDescriptor.
func (sd *PrivateDescriptor) writeTo(t *table) {
	t.row(0, "private_descriptor() {", nil)
	t.row(1, "splice_descriptor_tag", fmt.Sprintf("%#02x", sd.Tag()))
	t.row(1, "descriptor_length", sd.length())
	t.row(1, "identifier", fmt.Sprintf("%#08x, (%s)", sd.Identifier, sd.IdentifierString()))
	t.row(1, "private_bytes", fmt.Sprintf("%#0x", sd.PrivateBytes))
	t.row(0, "}", nil)
}
