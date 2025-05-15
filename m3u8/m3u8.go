// Package m3u8 implements reading and writing of m3u8 playlists
// used in HTTP Live Streaming (HLS) as specified in RFC 8216.
package m3u8

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/untangledco/streaming/scte35"
)

const MimeType string = "application/vnd.apple.mpegurl"

const RFC3339Milli string = "2006-01-02T15:04:05.999Z07:00"

type Playlist struct {
	Version             int
	Segments            []Segment
	IndependentSegments bool
	Start               *StartPoint

	// Media playlist
	// RFC 8216, 4.4.3.1
	TargetDuration        time.Duration
	Sequence              int
	DiscontinuitySequence int
	End                   bool
	Type                  PlaylistType
	IFramesOnly           bool

	// Master playlist
	Media       []Rendition
	Variants    []Variant
	SessionData []SessionData
	SessionKey  *Key
}

type Segment struct {
	URI string
	// Duration of this specific segment from the #EXTINF tag.
	Duration time.Duration
	// Indicates this segment holds a subset of the segment point to by URI.
	// Range is the length of the subsegment from from the #EXT-X-BYTERANGE tag.
	Range ByteRange
	// If true, the preceding segment and the following segment
	// are discontinuous. For example, this segment is part of a
	// commercial break.
	Discontinuity bool
	// Holds information on how to decrypt this segment.
	// If nil, the segment is not encrypted.
	Key       *Key
	Map       *Map
	DateTime  time.Time
	DateRange *DateRange
}

// Key represents the EXT-X-KEY tag specified in RFC 8216 seciton 4.3.2.3.
// A Key specifies how to decrypt encrypted playlist segments.
type Key struct {
	Method EncryptMethod
	// A URI pointing to instructions on how to obtain the key.
	URI    string
	Format string
	// An optional specification of the version of the key format
	// set in Format. The first value of the slice is the major
	// version; subsequent values are minor versions.
	FormatVersions []uint32
	// IV is a 128-bit unsigned integer holding the key's
	// initialisation vector.
	IV [16]byte
}

func (k Key) String() string {
	var attrs []string
	attrs = append(attrs, fmt.Sprintf("METHOD=%s", k.Method))
	attrs = append(attrs, fmt.Sprintf("URI=%q", k.URI))
	attrs = append(attrs, fmt.Sprintf("IV=0x%s", hex.EncodeToString(k.IV[:])))
	if k.Format != "" {
		attrs = append(attrs, fmt.Sprintf("KEYFORMAT=%q", k.Format))
	}
	if k.FormatVersions != nil {
		ss := make([]string, len(k.FormatVersions))
		for i := range k.FormatVersions {
			ss[i] = strconv.Itoa(int(k.FormatVersions[i]))
		}
		attrs = append(attrs, fmt.Sprintf("KEYFORMATVERSIONS=%q", strings.Join(ss, "/")))
	}
	return tagKey + ":" + strings.Join(attrs, ",")
}

const defaultKeyFormat string = "identity"

type EncryptMethod uint8

const (
	EncryptMethodNone EncryptMethod = 0 + iota
	EncryptMethodAES128
	EncryptMethodSampleAES
	encryptMethodInvalid EncryptMethod = 255
)

func (m EncryptMethod) String() string {
	switch m {
	case EncryptMethodNone:
		return "NONE"
	case EncryptMethodAES128:
		return "AES-128"
	case EncryptMethodSampleAES:
		return "SAMPLE-AES"
	}
	return "invalid"
}

func parseEncryptMethod(s string) EncryptMethod {
	switch s {
	case EncryptMethodNone.String():
		return EncryptMethodNone
	case EncryptMethodAES128.String():
		return EncryptMethodAES128
	case EncryptMethodSampleAES.String():
		return EncryptMethodSampleAES
	}
	return encryptMethodInvalid
}

// Map represents the EXT-X-MAP tag.
// A Map informs of any byte sequences to initialise readers of
// some media formats.
type Map struct {
	URI       string
	ByteRange ByteRange
}

func (m Map) String() string {
	if m.ByteRange != [2]int{0, 0} {
		return fmt.Sprintf("%s:URI=%q,BYTERANGE=%s", tagMap, m.URI, m.ByteRange)
	}
	return fmt.Sprintf("%s:URI=%q", tagMap, m.URI)
}

// ByteRange represents...
// The first entry is an offset, the second...?
type ByteRange [2]int

func (r ByteRange) String() string {
	if r[1] == 0 {
		return strconv.Itoa(r[0])
	}
	return fmt.Sprintf("%d@%d", r[0], r[1])
}

type DateRange struct {
	ID       string
	Class    string
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Planned  time.Duration
	// value must be a string, float or hex sequence (int?)
	Custom     map[string]any
	CueCommand *scte35.Splice
	// Contains the first of the in/out cue pair. Command may be
	// TimeSignal or Insert, with OutOfNetwork set to true.
	CueOut *scte35.Splice
	// Contains the second of the cue in/out pair. The Command's
	// Type must match the "out" cue.
	CueIn     *scte35.Splice
	EndOnNext bool
}

type PlaylistType uint8

const (
	PlaylistNone PlaylistType = iota
	PlaylistEvent
	PlaylistVOD
)

func (t PlaylistType) String() string {
	switch t {
	case PlaylistEvent:
		return "EVENT"
	case PlaylistVOD:
		return "VOD"
	}
	return "invalid"
}

type StartPoint struct {
	Offset  float32
	Precise bool
}
