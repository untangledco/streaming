package wav

import (
	"io"
	"os"
	"reflect"
	"testing"
)

func TestDecodeEncode(t *testing.T) {
	source := make([]byte, headerLength+extensionLength)
	f, err := os.Open("test.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := io.ReadFull(f, source[:]); err != nil {
		t.Fatalf("copy header bytes: %v", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	file, err := ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	header := EncodeHeader(file.Header)
	if !reflect.DeepEqual(source, header) {
		t.Errorf("encode header: want %v, got %v", source, header)
	}
}
