package sdp

import (
	"bytes"
	"fmt"
	"strings"
)

func (s Session) MarshalText() ([]byte, error) {
	buf := &strings.Builder{}
	fmt.Fprintln(buf, "v=0")

	if s.Origin.Username == "" {
		s.Origin.Username = NoUsername
	}

	fmt.Fprintln(buf, s.Origin)

	fmt.Fprintf(buf, "s=%s\n", s.Name)

	if s.Info != "" {
		fmt.Fprintf(buf, "i=%s\n", s.Info)
	}
	if s.URI != nil {
		fmt.Fprintf(buf, "u=%s\n", s.URI)
	}
	if s.Email != nil {
		// Remove quotes from mail.Address.String() to be
		// identical to examples in RFC 8866.
		fmt.Fprintf(buf, "e=%s\n", strings.ReplaceAll(s.Email.String(), `"`, ""))
	}
	if s.Phone != "" {
		fmt.Fprintf(buf, "p=%s\n", s.Phone)
	}
	if s.Connection != nil {
		fmt.Fprintln(buf, s.Connection)
	}
	if s.Bandwidth != nil {
		fmt.Fprintln(buf, s.Bandwidth)
	}

	if s.Time[0].IsZero() {
		fmt.Fprintln(buf, "t=0 0")
	} else {
		fmt.Fprintf(buf, "t=%d %d\n", s.Time[0].Unix()+sinceTimeZero, s.Time[1].Unix()+sinceTimeZero)
	}

	if s.Repeat != nil {
		fmt.Fprintln(buf, s.Repeat)
	}

	if s.Adjustments != nil {
		adj := make([]string, len(s.Adjustments))
		for i := range s.Adjustments {
			adj[i] = s.Adjustments[i].String()
		}
		fmt.Fprintf(buf, "z=%s\n", strings.Join(adj, " "))
	}

	if s.Attributes != nil {
		fmt.Fprintf(buf, "a=%s\n", strings.Join(s.Attributes, " "))
	}

	for _, m := range s.Media {
		fmt.Fprintln(buf, m)
	}

	return bytes.TrimSpace([]byte(buf.String())), nil
}
