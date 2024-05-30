package m3u8

import "testing"

func TestWriteSessionData(t *testing.T) {
	v := SessionData{
		ID:       "1234",
		Value:    "5678",
		URI:      "hello/hi.json",
		Language: "indonesian",
	}
	want := `DATA-ID="1234",VALUE="5678",URI="hello/hi.json",LANGUAGE="indonesian"`
	if v.String() != want {
		t.Errorf("unexpected session data text")
		t.Log("got:", v.String())
		t.Log("want:", want)
	}
}
