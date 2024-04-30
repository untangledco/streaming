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

const (
	// PrivateCommandType is the splice_command_type for private_command()
	PrivateCommandType = 0xFF
)

// PrivateCommand provides a means to distribute user-defined commands using the
// SCTE 35 protocol. The first bit field in each user-defined command is a
// 32-bit identifier, unique for each participating vendor. Receiving equipment
// should skip any splice_info_section() messages containing private_command()
// structures with unknown identifiers.
type PrivateCommand struct {
	XMLName      xml.Name `xml:"http://www.scte.org/schemas/35 PrivateCommand" json:"-"`
	JSONType     uint32   `xml:"-" json:"type"`
	Identifier   uint32   `xml:"identifier,attr" json:"identifier"`
	PrivateBytes Bytes    `xml:",chardata" json:"privateBytes"`
}

// IdentifierString returns the identifier as a string.
func (cmd *PrivateCommand) IdentifierString() string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, cmd.Identifier)
	return string(b)
}

// Type returns the splice_command_type.
func (cmd *PrivateCommand) Type() uint32 {
	// ensure JSONType is set
	cmd.JSONType = PrivateCommandType
	return PrivateCommandType
}

// decode a binary private_command.
func (cmd *PrivateCommand) decode(b []byte) error {
	r := iobit.NewReader(b)

	cmd.Identifier = r.Uint32(32)
	cmd.PrivateBytes = r.LeftBytes()
	// LeftBytes doesnt advance position
	r.Skip(uint(len(cmd.PrivateBytes) * 8))

	if err := readerError(r); err != nil {
		return fmt.Errorf("private_command: %w", err)
	}
	return readerError(r)
}

// encode this private_command to binary.
func (cmd *PrivateCommand) encode() ([]byte, error) {
	buf := make([]byte, cmd.length())

	iow := iobit.NewWriter(buf)
	iow.PutUint32(32, cmd.Identifier)
	_, err := iow.Write(cmd.PrivateBytes)
	if err != nil {
		return buf, err
	}

	err = iow.Flush()
	return buf, err
}

// commandLength returns the splice_command_length.
func (cmd *PrivateCommand) length() int {
	length := 32                        // identifier
	length += len(cmd.PrivateBytes) * 8 // private_bytes
	return length / 8
}

// writeTo the given table.
func (cmd *PrivateCommand) writeTo(t *table) {
	t.row(0, "private_command() {", nil)
	t.row(1, "identifier", fmt.Sprintf("%#08x, (%s)", cmd.Identifier, cmd.IdentifierString()))
	t.row(1, "private_byte", fmt.Sprintf("%#0x", cmd.PrivateBytes))
	t.row(0, "}", nil)
}
