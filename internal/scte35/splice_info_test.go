package scte35

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestPackTier(t *testing.T) {
	want := maxTier - 1
	packed := packTier(want)
	got := uint16(packed[1]) | uint16(packed[0])<<8
	if got != want {
		t.Errorf("want packed tier %d, got %d", want, got)
	}
}

func diffInfo(a, b SpliceInfo) string {
	buf := &strings.Builder{}
	if a.SAPType != b.SAPType {
		fmt.Fprintf(buf, "SAP type = %s, %s\n", a.SAPType, b.SAPType)
	}
	if a.Cipher != b.Cipher {
		fmt.Fprintln(buf, "cipher differs")
		fmt.Fprintf(buf, "< %s\n> %s", a.Cipher, b.Cipher)
	}
	if a.PTSAdjustment != b.PTSAdjustment {
		fmt.Fprintln(buf, "pts adjustment = ", a.PTSAdjustment, b.PTSAdjustment)
	}
	if a.CWIndex != b.CWIndex {
		fmt.Fprintln(buf, "cw index differs")
		fmt.Fprintf(buf, "< %v\n> %v\n", a.CWIndex, b.CWIndex)
	}
	if a.Tier != b.Tier {
		fmt.Fprintln(buf, "tier differs")
		fmt.Fprintf(buf, "< %#x\n> %#x\n", a.Tier, b.Tier)
	}
	if !reflect.DeepEqual(a.Command, b.Command) {
		fmt.Fprintln(buf, "command = ", *a.Command, *b.Command)
	}
	if !reflect.DeepEqual(a.Descriptors, b.Descriptors) {
		if len(a.Descriptors) != len(b.Descriptors) {
			fmt.Fprintf(buf, "descriptor count = %d, %d\n", len(a.Descriptors), len(b.Descriptors))
			fmt.Fprintf(buf, "descriptors = %+v, %+v\n", a.Descriptors, b.Descriptors)
		} else {
			fmt.Fprintln(buf, "descriptors differ")
			for i := range a.Descriptors {
				fmt.Fprintln(buf, "descriptor", i)
				buf.WriteString(diffDescriptors(a.Descriptors[i], b.Descriptors[i]))
			}
		}
	}
	if a.CRC32 != b.CRC32 {
		fmt.Fprintln(buf, "crc32 = ", a.CRC32, b.CRC32)
	}
	return buf.String()
}

func diffDescriptors(a, b SpliceDescriptor) string {
	buf := &strings.Builder{}
	if a.Tag() != b.Tag() {
		fmt.Fprintln(buf, "tag differs")
		fmt.Fprintln(buf, "<", a.Tag())
		fmt.Fprintln(buf, ">", b.Tag())
	}
	if a.ID() != b.ID() {
		fmt.Fprintln(buf, "id differs")
		fmt.Fprintf(buf, "< %d\n> %d\n", a.ID(), b.ID())
	}
	if !reflect.DeepEqual(a.Data(), b.Data()) {
		fmt.Fprintln(buf, "data differs")
		fmt.Fprintf(buf, "< %v\n> %v\n", a.Data(), b.Data())
	}
	fmt.Fprintf(buf, "< %T %v\n> %T %v\n", a, a, b, b)
	return buf.String()
}

func TestDecodeSpliceInfo(t *testing.T) {
	for _, tt := range samples {
		t.Run(tt.name, func(t *testing.T) {
			b, err := base64.StdEncoding.DecodeString(tt.encoded)
			if err != nil {
				t.Fatal("decode example splice info:", err)
			}
			info, err := decodeSpliceInfo(b)
			if err != nil {
				t.Fatalf("decode splice info: %v", err)
			}

			// test each possible command
			if tt.want.Command.TimeSignal != nil {
				if *tt.want.Command.TimeSignal != *info.Command.TimeSignal {
					t.Errorf("want timesig %x, got %x", *tt.want.Command.TimeSignal, *info.Command.TimeSignal)
				}
			}
			if tt.want.Command.Insert != nil {
				want := *tt.want.Command.Insert
				got := *info.Command.Insert
				if !reflect.DeepEqual(want, got) {
					t.Errorf("info command: want %+v, got %+v", want, got)
					if *want.SpliceTime != *got.SpliceTime {
						t.Logf("want splice time %d, got %d", want.SpliceTime, got.SpliceTime)
					}
					if want.Duration.Duration != got.Duration.Duration {
						t.Logf("want break duration %d, got %d", want.Duration.Duration, got.Duration.Duration)
					}
				}
			}

			if !reflect.DeepEqual(tt.want, *info) {
				t.Errorf("decode splice info: want %+v, got %+v", tt.want, *info)
				t.Log(diffInfo(tt.want, *info))
			}
		})
	}
}

func TestEncodeSpliceInfo(t *testing.T) {
	for _, tt := range samples {
		t.Run(tt.name, func(t *testing.T) {
			b, err := encodeSpliceInfo(&tt.want)
			if err != nil {
				t.Fatal(err)
			}
			got := base64.StdEncoding.EncodeToString(b)
			bwant, err := base64.StdEncoding.DecodeString(tt.encoded)
			if err != nil {
				t.Fatal(err)
			}
			if tt.encoded != got {
				t.Errorf("expected encoded splice info differs from calculated")
				t.Logf("< %#x", bwant)
				t.Logf("> %#x", b)
			}
		})
	}
}
