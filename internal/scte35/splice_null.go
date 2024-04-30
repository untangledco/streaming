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
)

const (
	// SpliceNullType is the splice_command_type for splice_null()
	SpliceNullType = 0x00
)

// SpliceNull is the command is provided for extensibility of the standard. The
// splice_null() command allows a splice_info_table to be sent that can carry
// descriptors without having to send one of the other defined commands. This
// command may also be used as a “heartbeat message” for monitoring cue
// injection equipment integrity and link integrity.
type SpliceNull struct {
	XMLName  xml.Name `xml:"http://www.scte.org/schemas/35 SpliceNull" json:"-"`
	JSONType uint32   `xml:"-" json:"type"`
}

// Type returns the splice_command_type.
func (cmd *SpliceNull) Type() uint32 {
	// ensure JSONType is set
	cmd.JSONType = SpliceNullType
	return SpliceNullType
}

// decode a binary splice_null.
func (cmd *SpliceNull) decode(b []byte) error {
	if len(b) > 0 {
		return fmt.Errorf("splice_null: %w", ErrBufferOverflow)
	}
	return nil
}

// encode this splice_null to binary.
func (cmd *SpliceNull) encode() ([]byte, error) {
	var b []byte
	return b, nil
}

// commandLength returns the splice_command_length.
func (cmd *SpliceNull) length() int {
	return 0
}

// writeTo the given table.
func (cmd *SpliceNull) writeTo(t *table) {
	t.row(0, "splice_null() {}", nil)
}
