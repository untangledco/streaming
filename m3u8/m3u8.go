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
	Version  int
	Segments []Segment

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

	// Both Media and Master playlists.
	IndependentSegments bool
	Start               *StartPoint
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

const defaultKeyFormat string = "identity"

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
	return strings.Join(attrs, ",")
}

type EncryptMethod uint8

const (
	EncryptMethodNone EncryptMethod = 0 + iota
	EncryptMethodAES128
	EncryptMethodSampleAES
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

type Map struct {
	URI       string
	ByteRange ByteRange
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
	CueCommand *scte35.SpliceInfo
	// Contains the first of the in/out cue pair. Command may be
	// TimeSignal or Insert, with OutOfNetwork set to true.
	CueOut *scte35.SpliceInfo
	// Contains the second of the cue in/out pair. The Command's
	// Type must match the "out" cue.
	CueIn     *scte35.SpliceInfo
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

// Rendition represents a unique rendition as described by a single
// EXT-X-MEDIA tag.
type Rendition struct {
	Type            MediaType
	URI             string
	Group           string
	Language        string
	AssocLanguage   string
	Name            string
	Default         bool
	AutoSelect      bool
	Forced          bool
	InstreamID      *CCInfo
	Characteristics []string
	Channels        []string
}

type MediaType uint8

const (
	MediaAudio MediaType = 0 + iota
	MediaVideo
	MediaSubtitles
	MediaClosedCaptions
)

func (t MediaType) String() string {
	switch t {
	case MediaAudio:
		return "AUDIO"
	case MediaVideo:
		return "VIDEO"
	case MediaSubtitles:
		return "SUBTITLES"
	case MediaClosedCaptions:
		return "CLOSED-CAPTIONS"
	}
	return "invalid"
}

type CCInfo struct {
	ID      int
	Service bool
}

func (info *CCInfo) String() string {
	if info.Service {
		return fmt.Sprintf("%s%d", "SERVICE", info.ID)
	}
	return fmt.Sprintf("%s%d", "CC", info.ID)
}

const (
	CharacteristicTranscribesDialog      = "public.accessibility.transcribes-spoken-dialog"
	CharacteristicDescribesMusicAndSound = "public.accessibility.transcribes-spoken-dialog"
	ChractersticEasyToRead               = "public.easy-to-read"
	CharacteristicDescribesVideo         = "public.accessibility.describes-video"
)

// EXT-X-STREAM-INF 4.3.4.2
type Variant struct {
	URI              string
	Bandwidth        int
	AverageBandwidth int
	Codecs           []string
	Resolution       [2]int
	FrameRate        float32
	HDCP             HDCPLevel
	Audio            string
	Video            string
	Subtitles        string
	// May be NoClosedCaptions or the empty string to indicate
	// absence of closed captions.
	ClosedCaptions string
}

// NoClosedCaptions may be the value for Variant.ClosedCaptions to
// explicitly indicate that no closed captions are available for the
// Variant.
const NoClosedCaptions string = "NONE"

type HDCPLevel uint8

const (
	HDCPNone HDCPLevel = 0 + iota
	HDCPType0
	HDCPType1
)

func (l HDCPLevel) String() string {
	switch l {
	case HDCPNone:
		return "NONE"
	case HDCPType0:
		return "TYPE-0"
	case HDCPType1:
		return "TYPE-1"
	}
	return "unknown"
}

// IFrameInfo represents the EXT-X-I-FRAME-STREAM-INF tag.
// It has the same structure as Variant, but the following fields should be unset:
// - FrameRate
// - Audio
// - Subtitles
// - ClosedCaptions
type IFrameInfo Variant

// SessionData represents the EXT-X-SESSION-DATA tag.
type SessionData struct {
	ID       string
	Value    string
	URI      string
	Language string
}
