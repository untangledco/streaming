/*
Package cmcd provides types and functions for exchanging
Common Media Client Data (CMCD) as specified in CTA-5004.

The typical use case for servers is to read client playback
information from a HTTP GET request, then relay the information to a
database for later analysis.
For instance, clients sending CMCD information as a query parameter
can be read with ParseInfo.

	func (srv *Server) ServeSegment(w http.ResponseWriter, req *http.Request) {
		v := req.URL.Query()
		var info cmcd.Info
		if v.Has("CMCD") {
			info, err := cmcd.ParseInfo(v.Get("CMCD"))
			if err != nil {
				log.Println("parse cmcd info: %v: ignoring", err)
			}
			relayToMetricStore(&info)
		}
		// serve response...
	}
*/
package cmcd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	HeaderRequest = "CMCD-Request"
	HeaderObject  = "CMCD-Object"
	HeaderStatus  = "CMCD-Status"
	HeaderSession = "CMCD-Session"
)

type Info struct {
	Request
	Object
	Status
	Session
	// Holds custom attributes as either a string, integer or
	// boolean.
	Custom map[string]any
}

func (info Info) Encode() string {
	ss := make([]string, 4)
	ss[0] = info.Request.Encode()
	ss[1] = info.Object.Encode()
	ss[2] = info.Status.Encode()
	ss[3] = info.Session.Encode()
	if info.Custom != nil && len(info.Custom) > 0 {
		for k, v := range info.Custom {
			switch v.(type) {
			case string:
				ss = append(ss, fmt.Sprintf("%s=%q", k, v))
			case int:
				ss = append(ss, fmt.Sprintf("%s=%d", k, v))
			case bool:
				ss = append(ss, k)
			default:
				ss = append(ss, fmt.Sprintf("%s=%q", k, v))
			}
		}
	}

	var noEmpty []string
	for _, s := range ss {
		if s == "" {
			continue
		}
		noEmpty = append(noEmpty, s)
	}
	return strings.Join(noEmpty, ",")
}

// ParseInfo returns the Info encoded in the string s.
// Typical usage is to parse a URL containing the "CMCD" query parameter,
// then pass the corresponding value to ParseInfo.
// See ExampleParseInfo.
func ParseInfo(s string) (Info, error) {
	return parseInfo(lex(s))
}

func ExtractInfo(header http.Header) (Info, error) {
	var fields []string
	fields = append(fields, header.Get(HeaderRequest))
	fields = append(fields, header.Get(HeaderObject))
	fields = append(fields, header.Get(HeaderStatus))
	fields = append(fields, header.Get(HeaderSession))
	tokens := lex(strings.Join(fields, ","))
	return parseInfo(tokens)
}

// Request represents data relating to the client's...
type Request struct {
	// Playback duration of the requested content. When encoded,
	// values are rounded to the nearest 100 milliseconds.
	BufLength time.Duration
	// Time limit to receive a response to the request before the
	// client may experience playback problems.
	// When encoded, values are rounded to the nearest 100 milliseconds.
	Deadline time.Duration
	// Kilobits per second between client and server, as measured by the client.
	Throughput int
	// Relative path of the next request.
	Next string
	// Byte range of the next request.
	NextRange Range
	// If true, a response is needed urgently as playback may be
	// starting, seeking, or the client has an empty playback buffer.
	Startup bool
}

type Range [2]int

func (r Range) String() string {
	if r[1] < 0 {
		return fmt.Sprintf("%d-", r[0])
	}
	return fmt.Sprintf("%d-%d", r[0], r[1])
}

type Object struct {
	// Encoded bitrate, in kilobits per second.
	Bitrate int
	// Playback duration. When encoded, values are rounded to the
	// nearest millisecond.
	Duration time.Duration
	// Media type, such as audio or video.
	Type ObjectType
	// The client's highest allowed bitrate, in kilobits per second.
	TopBitrate int
}

type ObjectType string

const (
	ObjTypeText      ObjectType = "m"
	ObjTypeAudio     ObjectType = "a"
	ObjTypeVideo     ObjectType = "v"
	ObjTypeAV        ObjectType = "av"
	ObjTypeI         ObjectType = "i"
	ObjTypeCaption   ObjectType = "c"
	ObjTypeTimedText ObjectType = "tt"
	ObjTypeKey       ObjectType = "k"
	ObjTypeOther     ObjectType = "o"
)

func parseRange(s string) (Range, error) {
	offset, end, found := strings.Cut(s, "-")
	if !found {
		return Range{}, fmt.Errorf("parse next range request: missing range separator %q", "-")
	}
	off, err := strconv.Atoi(offset)
	if err != nil {
		return Range{}, fmt.Errorf("offset: %w", err)
	}
	e, err := strconv.Atoi(end)
	if err != nil {
		return Range{}, fmt.Errorf("end: %w", err)
	}
	return Range{off, e}, nil
}

type Status struct {
	Starved       bool
	MaxThroughput int
}

type Session struct {
	// A GUID uniquely identifying the session, no longer than 64
	// characters.
	ID string
	// Type of the stream. If false, all playback segments are
	// available. Otherwise, the stream is considered live, and
	// segments become available over time.
	Live bool
	// A unique identifier of the client's requested content, no
	// longer than 64 characters.
	ContentID string
	// The playback rate of the content.
	PlayRate PlayRate
	// The format of the stream, such as HLS or MPEG-DASH.
	Format  StreamFormat
	version int
}

type PlayRate uint8

const (
	Stopped PlayRate = iota
	RealTime
	DoubleTime
)

type StreamFormat byte

const (
	FormatDASH   StreamFormat = 'd'
	FormatHLS    StreamFormat = 'h'
	FormatSmooth StreamFormat = 's'
	FormatOther  StreamFormat = 'o'
)

func (c StreamFormat) String() string {
	return fmt.Sprintf("%c", c)
}
