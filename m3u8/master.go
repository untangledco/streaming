package m3u8

import (
	"fmt"
	"strings"
)

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

func (r Rendition) String() string {
	var attrs []string
	attrs = append(attrs, fmt.Sprintf("NAME=%q", r.Name))
	attrs = append(attrs, fmt.Sprintf("TYPE=%s", r.Type))
	if r.URI != "" {
		attrs = append(attrs, fmt.Sprintf("URI=%q", r.URI))
	}
	attrs = append(attrs, fmt.Sprintf("GROUP-ID=%q", r.Group))
	if r.Language != "" {
		attrs = append(attrs, fmt.Sprintf("LANGUAGE=%q", r.Language))
	}
	if r.AssocLanguage != "" {
		attrs = append(attrs, fmt.Sprintf("ASSOC-LANGUAGE=%q", r.AssocLanguage))
	}
	if r.Default {
		attrs = append(attrs, "DEFAULT=YES")
	}
	if r.AutoSelect {
		attrs = append(attrs, "AUTOSELECT=YES")
	}
	if r.Forced {
		attrs = append(attrs, "FORCED=YES")
	}
	if r.Type == MediaClosedCaptions && r.InstreamID != nil {
		attrs = append(attrs, fmt.Sprintf("INSTREAM-ID=%q", r.InstreamID))
	}
	if len(r.Characteristics) > 0 {
		chars := strings.Join(r.Characteristics, ",")
		attrs = append(attrs, fmt.Sprintf("CHARACTERISTICS=%q", chars))
	}
	if len(r.Channels) > 0 {
		channels := strings.Join(r.Channels, "/")
		attrs = append(attrs, fmt.Sprintf("CHANNELS=%q", channels))
	}
	return tagRendition + ":" + strings.Join(attrs, ",")
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

// Variant represents the EXT-X-STREAM-INF tag specified in RFC 8216 section 4.3.4.2.
// The Audio, Video, Subtitles and ClosedCaptions fields each name a
// [Rendition] which can be combined to play the presentation.
type Variant struct {
	// Points to a media playlist carrying a rendition of this stream.
	URI string
	// Peak segment bitrate in bits per second.
	// This must be set.
	Bandwidth int
	// Average segment bitrate in bits per second.
	AverageBandwidth int

	// Codecs present in the stream in the format defined in ISO
	// Base Media File Format Name Space RFC6381 section 3.3.
	// For example: []string{"mp4a.40.2", "avc1.64001f"}
	Codecs []string

	// The width and height of the video in number of pixels.
	Resolution [2]int

	// Maximum video frame rate, in frames per second. Values are
	// rounded to 3 decimal places.
	FrameRate float32

	// Indicates if the stream is protected by High-bandwidth
	// Digital Content Protection (HDCP). The zero value
	// (HDCPLevelNone) indicates no output copy protection is
	// required for playing.
	HDCP      HDCPLevel

	// Each remaining field identifies a matching rendition of
	// the same type in the playlist. The match has its Group set to
	// the same value, and Type set to corresponding MediaType.
	// The empty string indicates there is no matching rendition in the playlist.
	//
	// For example, if Audio is set to "music", then a matching rendition
	// would be:
	//
	// 	Rendition{
	// 		Type: MediaAudio,
	//		Group: "music",
	//		...
	// 	}
	//
	// Similarly a matching rendition with Subtitles set to "sub1" would be:
	//
	// 	Rendition{
	//		Type: MediaSubtitles,
	//		Group: "sub1",
	//		...
	//	}
	Audio     string
	Video     string
	Subtitles string
	ClosedCaptions string // May be NoClosedCaptions to explicitly signal no rendition.
}

func (v Variant) String() string {
	var attrs []string
	attrs = append(attrs, fmt.Sprintf("BANDWIDTH=%d", v.Bandwidth))
	if v.AverageBandwidth > 0 {
		attrs = append(attrs, fmt.Sprintf("AVERAGE-BANDWIDTH=%d,", v.AverageBandwidth))
	}
	if len(v.Codecs) > 0 {
		attrs = append(attrs, fmt.Sprintf("CODECS=%q", strings.Join(v.Codecs, ",")))
	}
	if v.Resolution != [2]int{0, 0} {
		attrs = append(attrs, fmt.Sprintf("RESOLUTION=%dx%d", v.Resolution[0], v.Resolution[1]))
	}
	if v.FrameRate > 0 {
		attrs = append(attrs, fmt.Sprintf("FRAME-RATE=%.03f", v.FrameRate))
	}
	if v.HDCP != HDCPNone {
		attrs = append(attrs, fmt.Sprintf("HDCP-LEVEL=%s", v.HDCP))
	}
	if v.Audio != "" {
		attrs = append(attrs, fmt.Sprintf("AUDIO=%q", v.Audio))
	}
	if v.Video != "" {
		attrs = append(attrs, fmt.Sprintf("VIDEO=%q", v.Video))
	}
	if v.Subtitles != "" {
		attrs = append(attrs, fmt.Sprintf("SUBTITLES=%q", v.Subtitles))
	}
	if v.ClosedCaptions != "" && v.ClosedCaptions != NoClosedCaptions {
		attrs = append(attrs, fmt.Sprintf("CLOSED-CAPTIONS=%q", v.ClosedCaptions))
	}
	return fmt.Sprintf("%s:%s\n%s", tagVariant, strings.Join(attrs, ","), v.URI)
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
	ID       string // This attribute is REQUIRED
	Value    string // MUST contain either a VALUE or URI attribute, but not both
	URI      string
	Language string // This attribute is OPTIONAL.
}

func (sd SessionData) String() string {
	var attrs []string
	attrs = append(attrs, fmt.Sprintf("DATA-ID=%q", sd.ID))
	if sd.Value != "" {
		attrs = append(attrs, fmt.Sprintf("VALUE=%q", sd.Value))
	}
	if sd.URI != "" {
		attrs = append(attrs, fmt.Sprintf("URI=%q", sd.URI))
	}
	if sd.Language != "" {
		attrs = append(attrs, fmt.Sprintf("LANGUAGE=%q", sd.Language))
	}
	return tagSessionData + ":" + strings.Join(attrs, ",")
}
