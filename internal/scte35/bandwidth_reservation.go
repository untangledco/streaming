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
	// BandwidthReservationType is the splice_command_type for
	// bandwidth_reservation()
	BandwidthReservationType = 0x07
)

// BandwidthReservation command is provided for reserving bandwidth in a
// multiplex. A typical usage would be in a satellite delivery system that
// requires packets of a certain PID to always be present at the intended
// repetition rate to guarantee a certain bandwidth for that PID. This message
// differs from a splice_null() command so that it can easily be handled in a
// unique way by receiving equipment (i.e. removed from the multiplex by a
// satellite receiver). If a descriptor is sent with this command, it can not be
// expected that it will be carried through the entire transmission chain and it
// should be a private descriptor that is utilized only by the bandwidth
// reservation process.
type BandwidthReservation struct {
	XMLName  xml.Name `xml:"http://www.scte.org/schemas/35 BandwidthReservation" json:"-"`
	JSONType uint32   `xml:"-" json:"type"`
}

// Type returns the splice_command_type.
func (cmd *BandwidthReservation) Type() uint32 {
	// ensure JSONType is set
	cmd.JSONType = BandwidthReservationType
	return BandwidthReservationType
}

// decode a binary bandwidth_reservation.
func (cmd *BandwidthReservation) decode(b []byte) error {
	if len(b) > 0 {
		return fmt.Errorf("bandwidth_reservation: %w", ErrBufferOverflow)
	}
	return nil
}

// encode this bandwidth_reservation to binary.
func (cmd *BandwidthReservation) encode() ([]byte, error) {
	return nil, nil
}

// commandLength returns the splice_command_length
func (cmd *BandwidthReservation) length() int {
	return 0
}

// writeTo the given table.
func (cmd *BandwidthReservation) writeTo(t *table) {
	t.row(0, "bandwidth_reservation() {}", nil)
}
