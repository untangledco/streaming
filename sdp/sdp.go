// Package sdp implements a decoder and encoder for
// Session Description Protocol formatted data as specified in
// RFC 8866.
package sdp

import (
	"bufio"
	"fmt"
	"io"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
)

type Session struct {
	Version int
	Origin  Origin
	Name    string

	Info    string
	URI     *url.URL
	Email   *mail.Address
	Phone   string
	// TODO(otl): add rest of fields
}

type Origin struct {
	Username    string
	ID          int
	Version     int
	Network     string // TODO(otl): only "IN" is valid... so int type?
	AddressType string // TODO(otl): only "IP4", "IP6" valid... new int type?
	Address     string
}

func ReadSession(rd io.Reader) (*Session, error) {
	sc := bufio.NewScanner(rd)
	next := "v"
	var session Session
	for sc.Scan() {
		if sc.Text() == "" {
			continue // TODO(otl): empty lines allowed?
		}
		k, v, found := strings.Cut(sc.Text(), "=")
		if !found {
			return nil, fmt.Errorf("parse field %q: missing %q", next, "=")
		}
		if k != next {
			return nil, fmt.Errorf("expected field %q, found %q", next, k)
		}
		switch k {
		case "v":
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("parse version: %w", err)
			}
			if i != 0 {
				return nil, fmt.Errorf("parse version: unsupported version %d", i)
			}
			session.Version = i
			next = "o"
		case "o":
			o, err := parseOrigin(v)
			if err != nil {
				return nil, fmt.Errorf("parse origin: %w", err)
			}
			session.Origin = o
			next = "s"
		case "s":
			if v == "" {
				return nil, fmt.Errorf("empty name")
			}
			session.Name = v
		case "i":
			session.Info = v
			// TODO(otl): what do we expect next?
		case "u":
			u, err := url.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("parse uri: %w", err)
			}
			session.URI = u
			// TODO(otl): what do we expect next?
		case "e":
			addr, err := parseEmail(v)
			if err != nil {
				return nil, fmt.Errorf("parse email: %w", err)
			}
			session.Email = addr
			// TODO(otl): what do we expect next?
		default:
			return &session, fmt.Errorf("unsupported field %q", k)
		}
	}
	return &session, sc.Err()
}

func parseOrigin(line string) (Origin, error) {
	fields := strings.Fields(line)
	if len(fields) != 6 {
		return Origin{}, fmt.Errorf("need %d fields but only have %d", 6, len(fields))
	}
	o := Origin{Username: fields[0]}
	var err error
	o.ID, err = strconv.Atoi(fields[1])
	if err != nil {
		return o, fmt.Errorf("parse session id: %w", err)
	}
	o.Version, err = strconv.Atoi(fields[2])
	if err != nil {
		return o, fmt.Errorf("parse version: %w", err)
	}
	o.Network = fields[3]
	o.AddressType = fields[4]
	o.Address = fields[5]
	return o, nil
}

// parseEmail returns the parsed email address from s.
// Addresses should be in RFC 5322 form, for instance
// "Oliver Lowe <o@olowe.co>" or just "o@olowe.co".
// They can also be in the form detailed in the SDP specification,
// for instance
// "Oliver Lowe (o@olowe.co)".
func parseEmail(s string) (*mail.Address, error) {
	// Oliver Lowe (o@olowe.co) to RFC 5322 format
	// Oliver Lowe <o@olowe.co>
	s = strings.ReplaceAll(s, "(", "<")
	s = strings.ReplaceAll(s, ")", ">")
	return mail.ParseAddress(s)
}
