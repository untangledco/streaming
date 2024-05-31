package m3u8

import "testing"

func TestWriteSessionData(t *testing.T) {
	var cases = []struct {
		name string
		sd   SessionData
		want string
	}{
		{
			name: "all attributes set",
			sd:   SessionData{ID: "1234", Value: "5678", URI: "hello/hi.json", Language: "indonesian"},
			want: `DATA-ID="1234",VALUE="5678",URI="hello/hi.json",LANGUAGE="indonesian"`,
		},
		{
			name: "language is optional",
			sd:   SessionData{ID: "1234", Value: "5678", URI: "hello/hi.json"},
			want: `DATA-ID="1234",VALUE="5678",URI="hello/hi.json"`,
		},
		{
			name: "required attributes set(id, value or uri)",
			sd:   SessionData{ID: "1234", Value: "5678"},
			want: `DATA-ID="1234",VALUE="5678"`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sd.String() != tt.want {
				t.Errorf("unexpected session data text")
				t.Log("got:", tt.sd.String())
				t.Log("want:", tt.want)
			}
		})
	}
}
