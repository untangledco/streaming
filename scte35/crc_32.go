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

import "hash/crc32"

var crctab = makeCRC32Table(crc32PolyNormal)

// The reverse of crc32.IEEE, from
// https://en.wikipedia.org/wiki/Cyclic_redundancy_check#Polynomial_representations
const crc32PolyNormal = 0x04C11DB7

// makeCRC32Table generates CRC32/BZIP2 table using poly.
func makeCRC32Table(poly uint32) crc32.Table {
	var tab crc32.Table
	for i := range tab {
		crc := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if crc&0x80000000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc = crc << 1
			}
		}
		tab[i] = crc
	}
	return tab
}

// checksum calculates the checksum for the given bytes using tab.
func checksum(b []byte, tab *crc32.Table) uint32 {
	crc := int32(-1)
	for i := range b {
		crc = (crc << 8) ^ int32(tab[((crc>>24)^int32(b[i]))&0xFF])
	}
	return uint32(crc)
}
