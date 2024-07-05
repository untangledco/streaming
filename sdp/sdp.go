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
	Origin Origin
	Name   string

	Info      string
	URI       *url.URL
	Email     *mail.Address
	Phone     string
	Bandwidth *Bandwidth
	Media     []Media
	// TODO(otl): add rest of fields
}

type Origin struct {
	Username    string
	ID          int
	Version     int
	AddressType string // TODO(otl): only "IP4", "IP6" valid... new int type?
	Address     string // IPv4, IPv6 literal or a hostname
}

var fchars = [...]string{"i", "u", "e", "p", "c", "b", "t", "r", "z", "m"}

func ReadSession(rd io.Reader) (*Session, error) {
	session, sc, err := readSession(rd)
	if err != nil {
		return nil, err
	}

	// Time for optional fields. We keep a slice...
	// TODO(otl): document in plain language what's going on here.
	onext := fchars[:]
	for sc.Scan() {
		if sc.Text() == "" {
			return nil, fmt.Errorf("illegal empty line")
		}
		k, v, found := strings.Cut(sc.Text(), "=")
		if !found {
			return nil, fmt.Errorf("parse field %q: missing %q", k, "=")
		}

		var allowed bool
		for i := range onext {
			if onext[i] == k {
				allowed = true
			}
		}
		if !allowed {
			return nil, fmt.Errorf("unexpected field %q: expected one of %q", k, onext)
		}

		switch k {
		case "i":
			session.Info = v
			onext = fchars[1:]
		case "u":
			u, err := url.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("parse uri: %w", err)
			}
			session.URI = u
			onext = fchars[2:]
		case "e":
			addr, err := parseEmail(v)
			if err != nil {
				return nil, fmt.Errorf("parse email: %w", err)
			}
			session.Email = addr
			onext = fchars[3:]
		case "p":
			session.Phone = cleanPhone(v)
			onext = fchars[4:]
		case "b":
			bw, err := parseBandwidth(v)
			if err != nil {
				return nil, fmt.Errorf("parse bandwidth line %q: %w", v, err)
			}
			session.Bandwidth = &bw
			onext = fchars[5:]
		case "m":
			m, err := parseMedia(v)
			if err != nil {
				return nil, fmt.Errorf("parse media info from %q: %w", v, err)
			}
			session.Media = append(session.Media, m)
		}
	}

	return session, sc.Err()
}

func readSession(r io.Reader) (*Session, *bufio.Scanner, error) {
	sc := bufio.NewScanner(r)
	var session Session
	next := "v"
Loop:
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) == "" {
			return nil, nil, fmt.Errorf("illegal empty line")
		}
		k, v, found := strings.Cut(sc.Text(), "=")
		if !found {
			return nil, nil, fmt.Errorf("parse field %q: missing %q", next, "=")
		}
		if k != next {
			return nil, nil, fmt.Errorf("expected field %q, found %q", next, k)
		}
		switch k {
		case "v":
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, nil, fmt.Errorf("parse version: %w", err)
			}
			if i != 0 {
				return nil, nil, fmt.Errorf("unsupported version %d", i)
			}
			next = "o"
		case "o":
			o, err := parseOrigin(v)
			if err != nil {
				return nil, nil, fmt.Errorf("parse origin: %w", err)
			}
			session.Origin = o
			next = "s"
		case "s":
			if v == "" {
				return nil, nil, fmt.Errorf("empty name")
			}
			session.Name = v
			break Loop
		}
	}
	return &session, sc, sc.Err()
}

// cleanPhone returns the phone number in s stripped of "-" and space
// characters. Since "+1 617 555-6011" is semantically equal to
// "+16175556011", storing the number in the latter form lets us test for
// equality more easily.
func cleanPhone(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	return strings.ReplaceAll(s, "-", "")
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
	if fields[3] != "IN" {
		return o, fmt.Errorf("unknown network class %q", fields[3])
	}
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

type Bandwidth struct {
	Type    string
	Bitrate int // bits per second
}

func (b Bandwidth) String() string {
	// need kilobits per second as per section 5.8.
	return fmt.Sprintf("%s:%d", b.Type, b.Bitrate/1e3)
}

func parseBandwidth(s string) (Bandwidth, error) {
	t, b, ok := strings.Cut(s, ":")
	if !ok {
		return Bandwidth{}, fmt.Errorf("missing %s separator", ":")
	}
	// TODO(otl): check bandwith type is actually one specified in section 5.8.
	if t == "" {
		return Bandwidth{}, fmt.Errorf("missing bandwidth type")
	}
	kbps, err := strconv.Atoi(b)
	if err != nil {
		return Bandwidth{}, fmt.Errorf("parse bitrate: %w", err)
	}
	// convert to bits per second
	return Bandwidth{t, kbps * 1e3}, nil
}

type Media struct {
	Type      string // TODO(otl): new type mediaType?
	Port      int
	PortCount int
	Protocol  uint8
	Format    []string
}

const (
	ProtoUDP uint8 = iota
	ProtoRTP
	ProtoRTPSecure
	ProtoRTPSecureFeedback
)

func parseMedia(s string) (Media, error) {
	fields := strings.Fields(s)
	if len(fields) < 4 {
		return Media{}, fmt.Errorf("found %d fields, need at least %d", len(fields), 4)
	}
	m := Media{Type: fields[0]}

	p, n, found := strings.Cut(fields[1], "/")
	var err error
	m.Port, err = strconv.Atoi(p)
	if err != nil {
		return Media{}, fmt.Errorf("parse port: %w", err)
	}
	if found {
		m.PortCount, err = strconv.Atoi(n)
		if err != nil {
			return Media{}, fmt.Errorf("parse port count: %w", err)
		}
	}

	switch fields[2] {
	case "udp":
		m.Protocol = ProtoUDP
	case "RTP/AVP":
		m.Protocol = ProtoRTP
	case "RTP/SAVP":
		m.Protocol = ProtoRTPSecure
	case "RTP/SAVPF":
		m.Protocol = ProtoRTPSecureFeedback
	default:
		return Media{}, fmt.Errorf("unknown protocol %s", fields[2])
	}

	m.Format = fields[3:]
	return m, nil
}
