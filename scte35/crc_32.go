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
	"errors"
)

// ErrCRC32Invalid indicates that a splice_info_sections CRC_32 is
// invalid with respect to the binary payload.
var ErrCRC32Invalid = errors.New("CRC_32 not valid")

// crc32Table provides values for calculating an SCTE35 object's CRC32 value.
var crc32Table = makeCRC32Table()

// makeCRC32Table generates CRC32/BZIP2 table
func makeCRC32Table() [256]uint32 {
	const poly = 0x04C11DB7
	crctable := [256]uint32{}

	for i := range crctable {
		crc := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if crc&0x80000000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc = crc << 1
			}
		}
		crctable[i] = crc
	}
	return crctable
}

// calculateCRC32 takes an byte array Writer containing all of the bytes of a
// Splice Info Section object up to the CRC32 value, calculates a CRC32 value
// matching the bits for those existing bytes, and returns it.
func calculateCRC32(b []byte) uint32 {
	crc := int32(-1)
	for i := range b {
		crc = (crc << 8) ^ int32(crc32Table[((crc>>24)^int32(b[i]))&0xFF])
	}
	return uint32(crc)
}

// verifyCRC32 verifies the base-64 encoded binary string is valid by confirming
// the CRC_32 matches
func verifyCRC32(b []byte) error {
	if len(b) < 4 {
		return ErrCRC32Invalid
	}

	// crc32 is the last 4 bytes.
	payload := b[:len(b)-4]
	crc := binary.BigEndian.Uint32(b[len(b)-4:])
	if calculateCRC32(payload) != crc {
		return ErrCRC32Invalid
	}

	return nil
}
