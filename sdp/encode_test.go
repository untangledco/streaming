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
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	want := string(b)
	session, err := ReadSession(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if want != session.String() {
		t.Errorf("mismatched sdp text")
		t.Log("want", want)
		t.Log("got", session.String())
	}
}
