package scte35

var crctab = makeCRC32Table(crc32PolyNormal)

// The reverse of crc32.IEEE, from
// https://en.wikipedia.org/wiki/Cyclic_redundancy_check#Polynomial_representations
const crc32PolyNormal = 0x04C11DB7

// From Go's package compress/bzip2.
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in
// https://cs.opensource.google/go/go/+/master:LICENSE;bpv=0

// makeCRC32Table generates CRC32/BZIP2 table using poly.
func makeCRC32Table(poly uint32) [256]uint32 {
	var tab [256]uint32
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

func updateCRC(val uint32, b []byte) uint32 {
	crc := ^val
	for _, v := range b {
		crc = crctab[byte(crc>>24)^v] ^ (crc << 8)
	}
	return ^crc
}
