package scte35

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func diffInfo(a, b Splice) string {
	buf := &strings.Builder{}
	if a.SAPType != b.SAPType {
		fmt.Fprintln(buf, "SAP type differs")
		fmt.Fprintf(buf, "< %s\n> %s\n", a.SAPType, b.SAPType)
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
	for i := range a.Descriptors {
		if !reflect.DeepEqual(a.Descriptors[i], b.Descriptors[i]) {
			buf.WriteString(diffDescriptors(a.Descriptors[i], b.Descriptors[i]))
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

func TestDecode(t *testing.T) {
	for _, tt := range samples {
		t.Run(tt.name, func(t *testing.T) {
			b, err := base64.StdEncoding.DecodeString(tt.encoded)
			if err != nil {
				t.Fatal("decode example splice:", err)
			}
			splice, err := Decode(b)
			if err != nil {
				t.Fatalf("decode splice: %v", err)
			}

			// test each possible command
			if tt.want.Command.TimeSignal != nil {
				if *tt.want.Command.TimeSignal != *splice.Command.TimeSignal {
					t.Errorf("want timesig %x, got %x", *tt.want.Command.TimeSignal, *splice.Command.TimeSignal)
				}
			}
			if tt.want.Command.Insert != nil {
				want := *tt.want.Command.Insert
				got := *splice.Command.Insert
				if !reflect.DeepEqual(want, got) {
					t.Errorf("splice command: want %+v, got %+v", want, got)
					if *want.SpliceTime != *got.SpliceTime {
						t.Logf("want splice time %d, got %d", want.SpliceTime, got.SpliceTime)
					}
					if want.Duration.Duration != got.Duration.Duration {
						t.Logf("want break duration %d, got %d", want.Duration.Duration, got.Duration.Duration)
					}
				}
			}

			if !reflect.DeepEqual(tt.want, *splice) {
				t.Errorf("decode splice splice: want %+v, got %+v", tt.want, *splice)
				t.Log(diffInfo(tt.want, *splice))
			}
		})
	}

	// these messages are from github.com/futzu/SCTE-35_threefive/examples/hls/
	inserts := map[string]time.Duration{
		"/DAlAAAAAAAAAP/wFAUAAAABf+/+ANgNkv4AFJlwAAEBAQAA5xULLA==": 15 * time.Second,
		"/DAnAAAAAAAAAP/wBQb+AA27oAARAg9DVUVJAAAAAX+HCQA0AAE0xUZn": 10 * time.Second,
		"/DAnAAAAAAAAAP/wBQb+AGb/MAARAg9DVUVJAAAAAn+HCQA0AALMua1L": 75 * time.Second,
	}
	for s, dur := range inserts {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		splice, err := Decode(b)
		if err != nil {
			t.Fatalf("decode splice splice: %v", err)
		}
		var got time.Duration
		switch splice.Command.Type {
		case TimeSignal:
			got = time.Duration(*splice.Command.TimeSignal/90000) * time.Second
		case SpliceInsert:
			got = time.Duration(splice.Command.Insert.Duration.Duration/90000) * time.Second
		default:
			t.Fatalf("no duration test supported for %s", splice.Command.Type)
		}
		if got != dur {
			t.Errorf("want %s, got %s", dur, got)
		}
	}
}

func TestEncode(t *testing.T) {
	for _, tt := range samples {
		t.Run(tt.name, func(t *testing.T) {
			b, err := Encode(&tt.want)
			if err != nil {
				t.Fatal(err)
			}
			bwant, err := base64.StdEncoding.DecodeString(tt.encoded)
			if err != nil {
				t.Fatal(err)
			}
			// If we're not encrypted, set the CWIndex to be
			// the same as desired; its value is now undefined
			// and should be ignored downstream. This lets our
			// test pass even if our test encoded value has a
			// different CWIndex set than what we encode.
			if !tt.want.Encrypted {
				b[9] = bwant[9]
			}
			got := base64.StdEncoding.EncodeToString(b)
			if tt.encoded != got {
				// as above, since the undefined CWIndex is encoded differently,
				// our checksum could be different.
				// Only error if the value of the message *without* the CRC32 is different.
				if tt.encoded[:len(tt.encoded)-7] != got[:len(got)-7] {
					t.Errorf("expected encoded splice differs from calculated")
				}
				t.Logf("< %#x", bwant)
				t.Logf("> %#x", b)
			}
		})
	}
}
