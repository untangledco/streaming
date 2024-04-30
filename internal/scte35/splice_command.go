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

// NewSpliceCommand returns the splice command appropriate for the given type.
func NewSpliceCommand(spliceCommandType uint32) SpliceCommand {
	switch spliceCommandType {
	case SpliceNullType:
		return &SpliceNull{}
	case SpliceScheduleType:
		return &SpliceSchedule{}
	case SpliceInsertType:
		return &SpliceInsert{}
	case TimeSignalType:
		return &TimeSignal{}
	case BandwidthReservationType:
		return &BandwidthReservation{}
	default:
		return &PrivateCommand{}
	}
}

// SpliceCommand is an interface for splice_command.
type SpliceCommand interface {
	Type() uint32
	decode(b []byte) error
	encode() ([]byte, error)
	length() int
	writeTo(t *table)
}

// decodeSpliceCommand decodes the supplied byte array into the desired
// splice_command_type.
func decodeSpliceCommand(spliceCommandType uint32, b []byte) (SpliceCommand, error) {
	cmd := NewSpliceCommand(spliceCommandType)
	if err := cmd.decode(b); err != nil {
		return cmd, err
	}
	return cmd, nil
}
