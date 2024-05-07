package scte35

import "testing"

func TestPackEncryption(t *testing.T) {
	type ptest struct {
		sis  SpliceInfo
		want uint8
	}
	var tests = []ptest{
		{
			sis:  SpliceInfo{Encrypted: true, EncryptionAlgorithm: DES_CBC},
			want: 0b10000100,
		},
		{
			sis:  SpliceInfo{Encrypted: true, EncryptionAlgorithm: TripleDES},
			want: 0b10000110,
		},
	}
	for _, tt := range tests {
		var b byte
		if tt.sis.Encrypted {
			b |= (1 << 7)
		}
		b |= byte(tt.sis.EncryptionAlgorithm) << 1
		if b != tt.want {
			t.Errorf("pack encryption info %v: got %08b, want %08b", tt.sis, b, tt.want)
		}
	}
}
