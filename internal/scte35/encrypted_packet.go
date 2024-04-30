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

const (
	// EncryptionAlgorithmNone is the encryption_algorithm for None.
	EncryptionAlgorithmNone = 0
	// EncryptionAlgorithmDESECB is the encryption_algorithm for DES - ECB Mode.
	EncryptionAlgorithmDESECB = 1
	// EncryptionAlgorithmDESCBC is the encryption_algorithm for DES - CBC Mode.
	EncryptionAlgorithmDESCBC = 2
	// EncryptionAlgorithmTripleDES is the encryption_algorithm for Triple DES
	// EDE3 - ECB Mode.
	EncryptionAlgorithmTripleDES = 3
)

// EncryptedPacket contains the encryption details if this payload has been
// encrypted.
type EncryptedPacket struct {
	EncryptionAlgorithm uint32 `xml:"encryptionAlgorithm,attr,omitempty" json:"encryptionAlgorithm,omitempty"`
	CWIndex             uint32 `xml:"cwIndex,attr,omitempty" json:"cwIndex,omitempty"`
}

// encryptionAlgorithmName returns the user-friendly encryption algorithm name
func (p *EncryptedPacket) encryptionAlgorithmName() string {
	if p == nil {
		return "No encryption"
	}

	switch p.EncryptionAlgorithm {
	case EncryptionAlgorithmNone:
		return "No encryption"
	case EncryptionAlgorithmDESECB:
		return "DES – ECB mode"
	case EncryptionAlgorithmDESCBC:
		return "DES – CBC mode"
	case EncryptionAlgorithmTripleDES:
		return "Triple DES EDE3 – ECB mode"
	default:
		return "User private"
	}
}
