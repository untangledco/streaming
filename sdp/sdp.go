// Package sdp implements encoding and decoding of
// Session Description Protocol formatted data as specified in
// RFC 8866.
package sdp

import (
	"bufio"
	"fmt"
	"io"
	"net/mail"
	"net/netip"
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
	Time [2]time.Time
	// Repeat points to a repetition cycle describing for how long
	// and when the session may reoccur.
	Repeat *Repeat
	// Adjustments holds any time adjustments that may occur, for
	// example daylight savings, throughout the period a repetition
	// cycle is active.
	Adjustments []TimeAdjustment
	Attributes  []string
	Media       []Media
}

const NoUsername string = "-"

// Origin represents the originator of the session as described in RFC 8866 section 5.2.
type Origin struct {
	// Username is a named identity on the originating host. If unset,
	// the encoded value will be NoUsername.
	Username string

	// ID is a globally unique identifier for the session.
	// The recommended value is a timestamp from Now().
	ID int

	// Version is a version number of the session. It should be
	// incremented each time the session is modified. The recommended
	// value is a timestamp from Now().
	Version int

	// Address is the originating address of the session.
	Address netip.Addr
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

	// skip IP4/IP6 in fields[4]; netip handles the IP version for us.
	addr, err := netip.ParseAddr(fields[5])
	if err != nil {
		return o, fmt.Errorf("parse address: %w", err)
	}
	o.Address = addr
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

const (
	BandwidthConferenceTotal = "CT"
	BandwidthAppSpecific     = "AS"
)

type Bandwidth struct {
	// Type describes the value of Bitrate, usually one of
	// BandwidthConferenceTotal or BandwidthAppSpecific.
	Type string
	// Bitrate is the measure of bits per second.
	Bitrate int
}

func (b Bandwidth) String() string {
	// need kilobits per second as per section 5.8.
	return fmt.Sprintf("b=%s:%d", b.Type, b.Bitrate/1e3)
}

func parseBandwidth(s string) (Bandwidth, error) {
	t, b, ok := strings.Cut(s, ":")
	if !ok {
		return Bandwidth{}, fmt.Errorf("missing %s separator", ":")
	}
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

// Media represents a media description.
type Media struct {
	Type      MediaType
	Port      int // IP port
	PortCount int // count of subsequent ports from Port
	Transport TransportProto
	// Format describes the media format. Interpretation of the
	// entries depends on the value of Transport. For example, if
	// Transport is ProtoRTP, Format contains RTP payload type
	// numbers. For more, see the <fmt> description in section 5.14
	// of RFC 8866.
	Format []string

	// Optional fields
	Title      string
	Connection *ConnInfo
	Bandwidth  *Bandwidth
	Attributes []string
}

func (m Media) String() string {
	buf := &strings.Builder{}
	if m.PortCount == 0 {
		fmt.Fprintf(buf, "m=%s %d %s %s\n", m.Type, m.Port, m.Transport, strings.Join(m.Format, " "))
	} else {
		fmt.Fprintf(buf, "m=%s %d/%d %s %s\n", m.Type, m.Port, m.PortCount, m.Transport, strings.Join(m.Format, " "))
	}

	if m.Title != "" {
		fmt.Fprintf(buf, "i=%s\n", m.Title)
	}
	if m.Connection != nil {
		fmt.Fprintln(buf, m.Connection)
	}
	if m.Bandwidth != nil {
		fmt.Fprintln(buf, m.Bandwidth)
	}
	if m.Attributes != nil {
		fmt.Fprintf(buf, "a=%s", strings.Join(m.Attributes, " "))
	}
	return strings.TrimSpace(buf.String())
}

type MediaType uint8

const (
	MediaTypeAudio MediaType = iota
	MediaTypeVideo
	MediaTypeText
	MediaTypeApplication
	MediaTypeMessage
	MediaTypeImage
)

func (t MediaType) String() string {
	switch t {
	case MediaTypeAudio:
		return "audio"
	case MediaTypeVideo:
		return "video"
	case MediaTypeText:
		return "text"
	case MediaTypeApplication:
		return "application"
	case MediaTypeMessage:
		return "message"
	case MediaTypeImage:
		return "image"
	}
	return "unknown"
}

func parseMediaType(s string) (MediaType, error) {
	switch s {
	case MediaTypeAudio.String():
		return MediaTypeAudio, nil
	case MediaTypeVideo.String():
		return MediaTypeVideo, nil
	case MediaTypeText.String():
		return MediaTypeText, nil
	case MediaTypeApplication.String():
		return MediaTypeApplication, nil
	case MediaTypeMessage.String():
		return MediaTypeMessage, nil
	case MediaTypeImage.String():
		return MediaTypeImage, nil
	}
	return 0, fmt.Errorf("unknown media type %s", s)
}

type TransportProto uint8

const (
	ProtoUDP TransportProto = iota
	ProtoRTP
	ProtoRTPSecure
	ProtoRTPSecureFeedback
)

func (tp TransportProto) String() string {
	switch tp {
	case ProtoUDP:
		return "udp"
	case ProtoRTP:
		return "RTP/AVP"
	case ProtoRTPSecure:
		return "RTP/SAVP"
	case ProtoRTPSecureFeedback:
		return "RTP/SAVPF"
	}
	return "unknown"
}

func parseMedia(s string) (Media, error) {
	fields := strings.Fields(s)
	if len(fields) < 4 {
		return Media{}, fmt.Errorf("found %d fields, need at least %d", len(fields), 4)
	}

	mtyp, err := parseMediaType(fields[0])
	if err != nil {
		return Media{}, fmt.Errorf("media type: %w", err)
	}
	m := Media{Type: mtyp}

	p, n, found := strings.Cut(fields[1], "/")
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
	case ProtoUDP.String():
		m.Transport = ProtoUDP
	case ProtoRTP.String():
		m.Transport = ProtoRTP
	case ProtoRTPSecure.String():
		m.Transport = ProtoRTPSecure
	case ProtoRTPSecureFeedback.String():
		m.Transport = ProtoRTPSecureFeedback
	default:
		return Media{}, fmt.Errorf("unknown protocol %s", fields[2])
	}

	m.Format = fields[3:]
	return m, nil
}

// ConnInfo represents connection information.
type ConnInfo struct {
	Address netip.Addr
	// TTL is the time-to-live of multicast packets.
	TTL uint8
	// Count is the number of subsequent IP addresses after
	// Address used in the session.
	Count int
}

func (c *ConnInfo) String() string {
	ipv := "IP6"
	if c.Address.Is4() {
		ipv = "IP4"
	}
	s := fmt.Sprintf("c=%s %s %s", "IN", ipv, c.Address)
	if c.Address.Is4() && c.TTL > 0 {
		s += fmt.Sprintf("/%d", c.TTL)
	}
	if c.Count > 0 {
		s += fmt.Sprintf("/%d", c.Count)
	}
	return s
}

func parseConnInfo(s string) (ConnInfo, error) {
	fields := strings.Fields(s)
	if len(fields) != 3 {
		return ConnInfo{}, fmt.Errorf("expected %d fields, got %d", 3, len(fields))
	}
	if fields[0] != "IN" {
		return ConnInfo{}, fmt.Errorf("unsupported class %q, expected IN", fields[0])
	}

	var conn ConnInfo
	if fields[1] != "IP4" && fields[1] != "IP6" {
		return conn, fmt.Errorf("unsupported network type %s", fields[2])
	}
	addr := strings.Split(fields[2], "/")
	var err error
	conn.Address, err = netip.ParseAddr(addr[0])
	if err != nil {
		return conn, fmt.Errorf("parse address: %w", err)
	}
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

	if conn.Address.Is4() {
		if subfields[0] < 0 || subfields[0] > 255 {
			return conn, fmt.Errorf("ttl: %d is outside uint8 range", subfields[0])
		}
		conn.TTL = uint8(subfields[0])
		if len(subfields) == 2 {
			conn.Count = subfields[1]
		}
	}

	if conn.Address.Is6() && len(subfields) > 1 {
		return conn, fmt.Errorf("parse address: only 1 subfield allowed, read %d", len(subfields))
	} else if conn.Address.Is6() && len(subfields) == 1 {
		conn.Count = subfields[0]
	}
	return conn, nil
}
