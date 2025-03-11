package sip

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
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

	Proto      string
	ProtoMajor int
	ProtoMinor int

	Header        textproto.MIMEHeader
	ContentLength int64
	Sequence      int

	Body io.Reader
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
	req.Proto = version
	req.ProtoMajor = 2

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
		for _, r := range s {
			if unicode.IsSpace(r) {
				return line, fmt.Errorf("illegal space character in field %d", i)
			}
		}
	}
	return [3]string{fields[0], fields[1], fields[2]}, nil
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
	Sequence      CommandSequence

	Body io.ReadCloser
}
