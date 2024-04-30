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
	// TimeSignalType is the splice_command_type for a time_signal SpliceCommand.
	TimeSignalType = 0x06
)

// NewTimeSignal constructs a new time_signal command with the
// given pts_time value
func NewTimeSignal(ptsTime uint64) *TimeSignal {
	return &TimeSignal{
		SpliceTime: SpliceTime{
			PTSTime: &ptsTime,
		},
	}
}

// TimeSignal provides a time synchronized data delivery mechanism. The syntax
// of the time_signal() allows for the synchronization of the information
// carried in this message with the System Time Clock (STC). The unique payload
// of the message is carried in the descriptor, however the syntax and transport
// capabilities afforded to splice_insert() messages are also afforded to the
// time_signal(). The carriage however can be in a different PID than that
// carrying the other cue messages used for signaling splice points.
type TimeSignal struct {
	XMLName    xml.Name   `xml:"http://www.scte.org/schemas/35 TimeSignal" json:"-"`
	JSONType   uint32     `xml:"-" json:"type"`
	SpliceTime SpliceTime `xml:"http://www.scte.org/schemas/35 SpliceTime" json:"spliceTime"`
}

// Type returns the splice_command_type.
func (cmd *TimeSignal) Type() uint32 {
	// ensure JSONType is set
	cmd.JSONType = TimeSignalType
	return TimeSignalType
}

// decode a binary time_signal
func (cmd *TimeSignal) decode(b []byte) error {
	r := iobit.NewReader(b)
	timeSpecifiedFlag := r.Bit()
	if timeSpecifiedFlag {
		r.Skip(6) // reserved
		ptsTime := r.Uint64(33)
		cmd.SpliceTime.PTSTime = &ptsTime
	} else {
		r.Skip(7) // reserved
	}

	if err := readerError(r); err != nil {
		return fmt.Errorf("%s: %w", cmd.XMLName, err)
	}
	return nil
}

// encode this time_signal as binary.
func (cmd *TimeSignal) encode() ([]byte, error) {
	buf := make([]byte, cmd.length())

	iow := iobit.NewWriter(buf)
	if cmd.timeSpecifiedFlag() {
		iow.PutBit(true)
		iow.PutUint32(6, Reserved) // reserved
		iow.PutUint64(33, *cmd.SpliceTime.PTSTime)
	} else {
		iow.PutBit(false)
		iow.PutUint32(7, Reserved) // reserved
	}

	err := iow.Flush()
	return buf, err
}

// commandLength returns the splice_command_length.
func (cmd *TimeSignal) length() int {
	length := 1 // time_specified_flag
	if cmd.timeSpecifiedFlag() {
		length += 6  // reserved
		length += 33 // pts_time
	} else {
		length += 7 // reserved
	}
	return length / 8
}

// writeTo the given table.
func (cmd *TimeSignal) writeTo(t *table) {
	t.row(0, "time_signal() {", nil)
	t.row(1, "time_specified_flag", cmd.timeSpecifiedFlag())
	if cmd.timeSpecifiedFlag() {
		t.row(1, "pts_time", cmd.SpliceTime.PTSTime)
	}
	t.row(0, "}", nil)
}

// timeSpecifiedFlag return the time_specified_flag.
func (cmd *TimeSignal) timeSpecifiedFlag() bool {
	return cmd != nil && cmd.SpliceTime.PTSTime != nil
}
