package sdp

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestWriteSession(t *testing.T) {
	f, err := os.Open("testdata/good.sdp")
	if err != nil {
		t.Fatal(err)
	}
	want, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	session, err := ReadSession(bytes.NewReader(want))
	if err != nil {
		t.Fatal(err)
	}
	got, err := session.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, got) {
		t.Errorf("mismatched sdp text")
		t.Log("want", string(want))
		t.Log("got", string(got))
	}
}
