// Copyright 2024 The Untangled Authors. Use of this source code is
// governed by the ISC license available in the LICENSE.ISC file.

package scte35

const (
	EncryptionAlgorithmNone uint32 = iota
	EncryptionAlgorithmDESECB
	EncryptionAlgorithmDESCBC
	EncryptionAlgorithmTripleDES
)

// EncryptedPacket represents payload encryption information.
type EncryptedPacket struct {
	// Specifies which cipher is used.
	// See SCTE 35 section 11.3.
	EncryptionAlgorithm uint32
	CWIndex             uint32
}

// cipher is a 6-bit field specifying the algorithm used to encrypt
// payloads as defined in SCTE 35 section 11.3.
type cipher uint8

const (
	cipherNone cipher = iota
	des_ECB           // SCTE 35 section 11.3.1
	des_CBC           // SCTE 35 section 11.3.2
	tripleDES         // SCTE 35 section 11.3.3
	reserved
	// Values 32 through 63 are available for "User private"
	// algorithms. See SCTE 35 section 11.3.4.
)

func (c cipher) String() string {
	switch c {
	case cipherNone:
		return "none"
	case des_ECB:
		return "DES – ECB mode"
	case des_CBC:
		return "DES – CBC mode"
	case tripleDES:
		return "Triple DES EDE3 – ECB mode"
	}
	if c >= reserved && c < 32 {
		return "reserved"
	}
	return "user private"
}
