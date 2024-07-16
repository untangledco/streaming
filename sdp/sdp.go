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
	"time"
)

type Session struct {
	Origin Origin
	Name   string

	Info       string
	URI        *url.URL
	Email      *mail.Address
	Phone      string
	Connection *ConnInfo
	Bandwidth  *Bandwidth
	// Time holds the start time and stop time of the Session, at
	// the first and second index respectively.
	Time       [2]time.Time
	Repeat     *Repeat
	Attributes []string
	Media      []Media
}

type Origin struct {
	Username    string
	ID          int
	Version     int
	AddressType string // TODO(otl): only "IP4", "IP6" valid... new int type?
	Address     string // IPv4, IPv6 literal or a hostname
}

func ReadSession(rd io.Reader) (*Session, error) {
	parser := &parser{Scanner: bufio.NewScanner(rd)}
	if err := parser.parse(); err != nil {
		return nil, fmt.Errorf("parse session: %w", err)
	}
	return &parser.session, nil
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
// They can also be in the form detailed in the SDP specification, for instance
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
	// Optional fields
	Title      string
	Connection *ConnInfo
	Bandwidth  *Bandwidth
	// TODO(otl): store as k, v pairs
	Attributes []string
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

type ConnInfo struct {
	Type    string // TODO(otl): only "IP4", "IP6" valid... new int type?
	Address string // IPv4, IPv6 literal or a hostname
	TTL     int    // time to live
	Count   int    // number of addresses after Address
}

func parseConnInfo(s string) (ConnInfo, error) {
	fields := strings.Fields(s)
	if len(fields) != 3 {
		return ConnInfo{}, fmt.Errorf("expected %d fields, got %d", 3, len(fields))
	}
	if fields[0] != "IN" {
		return ConnInfo{}, fmt.Errorf("unsupported class %q, expected IN", fields[0])
	}

	conn := ConnInfo{Type: fields[1]}
	if fields[1] != "IP4" && fields[1] != "IP6" {
		return conn, fmt.Errorf("unsupported network type %s", fields[2])
	}
	conn.Type = fields[1]
	addr := strings.Split(fields[2], "/")
	conn.Address = addr[0]
	if len(addr) == 1 {
		return conn, nil
	}

	subfields := make([]int, len(addr[1:]))
	for i := range subfields {
		var err error
		subfields[i], err = strconv.Atoi(addr[i+1])
		if err != nil {
			return conn, fmt.Errorf("parse address subfield %d: %w", i, err)
		}
	}

	if conn.Type == "IP4" && len(subfields) == 2 {
		conn.TTL = subfields[0]
		conn.Count = subfields[1]
	} else if conn.Type == "IP4" && len(subfields) == 1 {
		conn.TTL = subfields[0]
	}

	if conn.Type == "IP6" && len(subfields) > 1 {
		return conn, fmt.Errorf("parse address: only 1 subfield allowed, read %d", len(subfields))
	} else if conn.Type == "IP6" && len(subfields) == 1 {
		conn.Count = subfields[0]
	}
	return conn, nil
}
