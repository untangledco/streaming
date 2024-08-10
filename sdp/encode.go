package sdp

import (
	"fmt"
	"strings"
)

func (s Session) String() string {
	buf := &strings.Builder{}
	fmt.Fprintln(buf, "v=0")
	if s.Origin.Username == "" {
		s.Origin.Username = NoUsername
	}
	fmt.Fprintf(buf, "o=%s %d %d IN %s %s\n", s.Origin.Username, s.Origin.ID, s.Origin.Version, s.Origin.AddressType, s.Origin.Address)
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

	// TODO(otl): what about the invalid case where Time[0] is zero but Time[1] is not?
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

	return strings.TrimSpace(buf.String())
}
