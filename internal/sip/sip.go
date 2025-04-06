// Package sip ... SIP protocol as specified in RFC 3261.
package sip

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

const (
	MethodRegister = "REGISTER"
	MethodInvite   = "INVITE"
	MethodAck      = "ACK"
	MethodCancel   = "CANCEL"
	MethodBye      = "BYE"
	MethodOptions  = "OPTIONS"
)

const version = "SIP/2.0"

type Request struct {
	Method string
	URI    string

	Header        textproto.MIMEHeader
	ContentLength int64
	ContentType   string
	Sequence      int
	To            Address
	From          Address
	Via           Via

	Body io.Reader
}

type URI url.URL

func (u URI) String() string {
	return "<" + (*url.URL)(&u).String() + ">"
}

type Address struct {
	Name string
	URI
	Tag string
}

func (a Address) String() string {
	var tag string
	if a.Tag != "" {
		tag = ";tag=" + a.Tag
	}
	if a.Name != "" {
		return fmt.Sprintf("%s %s%s", a.Name, a.URI, tag)
	}
	return a.URI.String() + tag
}

func ParseAddress(s string) (Address, error) {
	s = strings.TrimSpace(s)

	// TODO(otl): we're parsing header parameters - should we generalise somewhere?
	// See section 20.
	before, tag, found := strings.Cut(s, ";")
	if found {
		if !strings.HasPrefix(tag, "tag=") {
			return Address{}, fmt.Errorf("bad tag: missing %q prefix", "tag=")
		}
		tag = tag[4:]
	}
	addr := Address{Tag: tag}

	// bare URI without angle brackets
	// e.g. "sip:test@example.com"
	u, err := url.Parse(before)
	if err == nil {
		addr.URI = URI(*u)
		return addr, nil
	}

	// URI without name
	// e.g. "<sip:test@example.com>"
	if strings.HasPrefix(before, "<") && strings.HasSuffix(before, ">") {
		trimmed := strings.Trim(before, "<>")
		u, err := url.Parse(trimmed)
		if err != nil {
			return addr, err
		}
		addr.URI = URI(*u)
		return addr, nil
	}

	i := strings.Index(before, "<")
	if i < 0 {
		return addr, fmt.Errorf("missing angle bracket after name")
	}
	j := strings.Index(before, ">")
	if j < 0 {
		return addr, fmt.Errorf("missing closing angle bracket")
	}
	addr.Name = strings.TrimSpace(before[:i])

	u, err = url.Parse(before[i+1 : j])
	if err != nil {
		return addr, fmt.Errorf("parse uri: %w", err)
	}
	addr.URI = URI(*u)
	return addr, nil
}

const magicViaCookie = "z9hG4bK"

const (
	TransportUDP int = iota
	TransportTCP
)

// Via represents the Via field in the header of requests.
type Via struct {
	// Transport indicates whether TCP or UDP should be used in
	// subsequent transactions.
	Transport int
	// Address is a hostname or IP address to which responses
	// should be sent.
	Address string
	// Branch uniquely identifies transactions from a particular user-agent.
	Branch string
}

func (v Via) String() string {
	tport := "unknown"
	switch v.Transport {
	case TransportUDP:
		tport = "UDP"
	case TransportTCP:
		tport = "TCP"
	}
	return fmt.Sprintf("SIP/2.0/%s %s;branch=%s%s", tport, v.Address, magicViaCookie, v.Branch)
}

func ReadRequest(r io.Reader) (*Request, error) {
	msg, err := readMessage(r)
	if err != nil {
		return nil, err
	}
	return parseRequest(msg)
}

func parseRequest(msg *message) (*Request, error) {
	var req Request
	req.Method = msg.startLine[0]
	req.URI = msg.startLine[1]
	if msg.startLine[2] != version {
		return &req, fmt.Errorf("unknown version %q", msg.startLine[2])
	}

	req.Header = msg.header
	if s := req.Header.Get("Content-Length"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			return &req, fmt.Errorf("parse content-length: %w", err)
		}
		req.ContentLength = int64(n)
	}
	req.Body = msg.body
	return &req, nil
}

func WriteRequest(w io.Writer, req *Request) (n int64, err error) {
	// section 8.1.1. We can set Max-Forwards automatically.
	required := []string{"CSeq", "Call-ID"}
	for _, s := range required {
		if req.Header.Get(s) == "" {
			return 0, fmt.Errorf("missing field %s in header", s)
		}
	}

	if req.To.URI.String() == "" {
		return 0, fmt.Errorf("empty uri in to header field")
	}
	if req.From.URI.String() == "" {
		return 0, fmt.Errorf("empty uri in from header field")
	}
	req.Header.Set("To", req.To.String())
	req.Header.Set("From", req.From.String())

	if req.Via.Address == "" {
		return 0, fmt.Errorf("empty address in via header field")
	} else if req.Via.Branch == "" {
		return 0, fmt.Errorf("empty branch in via header field")
	}

	req.Header.Set("Via", req.Via.String())
	if req.Header.Get("Max-Forwards") == "" {
		// TODO(otl): find section in RFC recommending 70.
		// section x.x.x
		req.Header.Set("Max-Forwards", strconv.Itoa(70))
	}
	if req.ContentLength > 0 {
		req.Header.Set("Content-Length", strconv.Itoa(int(req.ContentLength)))
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%s %s SIP/2.0\r\n", req.Method, req.URI)
	for k := range req.Header {
		for _, v := range req.Header.Values(k) {
			fmt.Fprintf(buf, "%s: %s\r\n", k, v)
		}
	}
	buf.WriteString("\r\n")
	n, err = io.Copy(w, buf)
	if err != nil {
		return n, err
	}

	if req.Body != nil {
		var nn int64
		nn, err = io.Copy(w, req.Body)
		n += nn
	}
	return n, err
}

type message struct {
	startLine [3]string
	header    textproto.MIMEHeader
	body      *bufio.Reader
}

func readMessage(rd io.Reader) (*message, error) {
	r := textproto.NewReader(bufio.NewReader(rd))
	line, err := r.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("read start line: %w", err)
	}
	sline, err := parseStartLine(line)
	if err != nil {
		return nil, fmt.Errorf("parse start line: %w", err)
	}

	header, err := r.ReadMIMEHeader()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	return &message{startLine: sline, header: header, body: r.R}, nil
}

// Request-Line  =  Method SP Request-URI SP SIP-Version CRLF
// Status-Line  =  SIP-Version SP Status-Code SP Reason-Phrase CRLF
func parseStartLine(text string) (line [3]string, err error) {
	fields := strings.Fields(text)
	if len(fields) != 3 {
		return line, fmt.Errorf("expected 3 fields, read %d", len(fields))
	}
	for i, s := range fields {
		if containsSpace(s) {
			return line, fmt.Errorf("illegal space character in field %d", i)
		}
	}
	return [3]string{fields[0], fields[1], fields[2]}, nil
}

func containsSpace(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

type CommandSequence struct {
	Number int
	Method string
}

type Response struct {
	Status     string
	StatusCode int

	Header        textproto.MIMEHeader
	ContentLength int64
	// Sequence      CommandSequence

	Body io.Reader
}

func parseResponse(msg *message) (*Response, error) {
	if msg.startLine[0] != version {
		return nil, fmt.Errorf("unknown version %s", msg.startLine[0])
	}

	var resp Response
	var err error
	resp.StatusCode, err = strconv.Atoi(msg.startLine[1])
	if err != nil {
		return nil, fmt.Errorf("bad status code %q: %v", msg.startLine[1], err)
	}
	resp.Status = msg.startLine[2]

	resp.Header = msg.header
	if s := resp.Header.Get("Content-Length"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			return &resp, fmt.Errorf("parse content-length: %w", err)
		}
		resp.ContentLength = int64(n)
	}
	resp.Body = msg.body
	return &resp, nil
}
