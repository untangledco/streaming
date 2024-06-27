package cair

import (
	"encoding/xml"
	"strings"
	"testing"
)

const tstatus string = `<Status>
  <Active Id="{XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}"/>
  <Cued Id="{XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}"/>
  <License State="Licensed|Not Licensed|Demo"/>
  <Output State="Normal|Black|Bypass|Clean"/>
  <Client Connected="y|n" Identity="IdentityString"/>
</Status>`

func TestStatus(t *testing.T) {
	var status Status
	if err := xml.NewDecoder(strings.NewReader(tstatus)).Decode(&status); err != nil {
		t.Fatal(err)
	}
	b, err := xml.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}
