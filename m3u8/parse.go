package m3u8

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	tagHead                = tagStart + "M3U"
	tagVersion             = "#EXT-X-VERSION"
	tagVariant             = "#EXT-X-STREAM-INF"
	tagRendition           = "#EXT-X-MEDIA"
	tagPlaylistType        = "#EXT-X-PLAYLIST-TYPE"        // RFC 8216, 4.4.3.5
	tagTargetDuration      = "#EXT-X-TARGETDURATION"       // RFC 8216, 4.4.3.1
	tagMediaSequence       = "#EXT-X-MEDIA-SEQUENCE"       // RFC 8216, 4.3.3.2
	tagEndList             = "#EXT-X-ENDLIST"              // RFC 8216, 4.4.3.4
	tagIndependentSegments = "#EXT-X-INDEPENDENT-SEGMENTS" // RFC 8216, 4.3.5.1
	tagSessionData         = "#EXT-X-SESSION-DATA"         // RFC 8216, 4.3.4.4
)

func Decode(rd io.Reader) (*Playlist, error) {
	lex := newLexer(rd)
	go lex.run()
	it := <-lex.items
	if it.typ == itemError {
		return nil, errors.New(it.val)
	}
	if it.typ != itemTag || it.val != tagHead {
		return nil, fmt.Errorf("expected head tag, got %q", it.val)
	}
	p := &Playlist{}
	var err error
	for it := range lex.items {
		switch it.typ {
		case itemError:
			return p, errors.New(it.val)
		case itemTag:
			switch it.val {
			case tagVersion:
				it = <-lex.items
				if p.Version != 0 {
					return p, fmt.Errorf("parse %s: playlist version already specified", it)
				}
				p.Version, err = strconv.Atoi(it.val)
				if err != nil {
					return p, fmt.Errorf("parse playlist version: %w", err)
				}
			case tagIndependentSegments:
				p.IndependentSegments = true
			case tagVariant:
				variant, err := parseVariant(lex.items)
				if err != nil {
					return p, fmt.Errorf("parse variant: %w", err)
				}
				p.Variants = append(p.Variants, *variant)
			case tagRendition:
				rend, err := parseRendition(lex.items)
				if err != nil {
					return p, fmt.Errorf("parse rendition: %w", err)
				}
				p.Media = append(p.Media, *rend)
			case tagPlaylistType:
				it = <-lex.items
				typ, err := parsePlaylistType(it)
				if err != nil {
					return p, fmt.Errorf("parse playlist type: %w", err)
				}
				p.Type = typ

			case tagTargetDuration:
				it = <-lex.items
				dur, err := parseTargetDuration(it)
				if err != nil {
					return p, fmt.Errorf("parse target duration: %w", err)
				}
				p.TargetDuration = dur

			case tagSegmentDuration, tagByteRange, tagKey:
				segment, err := parseSegment(lex.items, it)
				if err != nil {
					return p, fmt.Errorf("parse segment: %w", err)
				}
				p.Segments = append(p.Segments, *segment)

			case tagEndList:
				p.End = true
			case tagMediaSequence:
				it = <-lex.items
				seq, err := strconv.Atoi(it.val)
				if err != nil {
					return p, fmt.Errorf("parse media sequence: %w", err)
				}
				p.Sequence = seq
			}
		}
	}
	return p, nil
}

func parseVariant(items chan item) (*Variant, error) {
	var v Variant
	for it := range items {
		switch it.typ {
		case itemAttrName:
			attr := it
			it = <-items
			if it.typ != itemEquals {
				return nil, fmt.Errorf("missing equals after %s", attr)
			}
			switch attr.val {
			case "PROGRAM-ID", "NAME":
				// parsing PROGRAM-ID attribute unsupported; removed in HLS version 6
				// NAME is non-standard, should be set in Rendition.
			case "BANDWIDTH", "AVERAGE-BANDWIDTH":
				it = <-items
				if it.typ != itemNumber {
					return nil, fmt.Errorf("parse bandwidth attribute: unexpected %s", it)
				}
				n, err := strconv.Atoi(it.val)
				if err != nil {
					return nil, fmt.Errorf("parse bandwidth: %w", err)
				}
				if attr.val == "BANDWIDTH" {
					v.Bandwidth = n
				} else {
					v.AverageBandwidth = n
				}
			case "CODECS":
				it = <-items
				if it.typ != itemString {
					return nil, fmt.Errorf("parse codecs attribute: unexpected %s", it)
				}
				v.Codecs = strings.Split(strings.Trim(it.val, `"`), ",")
			case "RESOLUTION":
				it = <-items
				res, err := parseResolution(it.val)
				if err != nil {
					return nil, fmt.Errorf("parse resolution: %w", err)
				}
				v.Resolution = res
			case "FRAME-RATE":
				it = <-items
				if it.typ != itemNumber {
					return nil, fmt.Errorf("parse frame rate: unexpected %s", it)
				}
				n, err := strconv.ParseFloat(it.val, 32)
				if err != nil {
					return nil, fmt.Errorf("parse frame rate: %w", err)
				}
				v.FrameRate = float32(n)
			case "HDCP-LEVEL":
				it = <-items
				l, err := parseHDCPLevel(it.val)
				if err != nil {
					return nil, fmt.Errorf("parse HDCP level: %w", err)
				}
				v.HDCP = l
			case "AUDIO", "VIDEO", "SUBTITLES":
				name := attr.val
				it = <-items
				if it.typ != itemString {
					return nil, fmt.Errorf("parse %s: unexpected %s", name, it)
				}
				it.val = strings.Trim(it.val, `"`)
				if name == "AUDIO" {
					v.Audio = it.val
				} else if name == "VIDEO" {
					v.Video = it.val
				} else if name == "SUBTITLES" {
					v.Subtitles = it.val
				}
			case "CLOSED-CAPTIONS":
				it = <-items
				if it.typ != itemString {
					return nil, fmt.Errorf("parse closed-captions: unexpected %s", it)
				}
				v.ClosedCaptions = strings.Trim(it.val, `"`)
			default:
				return nil, fmt.Errorf("unknown attribute %s", attr.val)
			}
		case itemComma:
			continue
		case itemURL:
			v.URI = it.val
			return &v, nil
		}
	}
	return &v, nil
}

func parseResolution(s string) (res [2]int, err error) {
	x, y, found := strings.Cut(s, "x")
	if !found {
		return res, fmt.Errorf("missing x seperator")
	}
	res[0], err = strconv.Atoi(x)
	if err != nil {
		return res, fmt.Errorf("horizontal pixels: %v", err)
	}
	res[1], err = strconv.Atoi(y)
	if err != nil {
		return res, fmt.Errorf("vertical pixels: %v", err)
	}
	if res[0] < 0 || res[1] < 0 {
		return res, fmt.Errorf("negative dimensions")
	}
	return res, nil
}

func parseHDCPLevel(s string) (HDCPLevel, error) {
	switch s {
	case "NONE":
		return HDCPNone, nil
	case "TYPE-0":
		return HDCPType0, nil
	case "TYPE-1":
		return HDCPType1, nil
	}
	return 0, fmt.Errorf("unknown HDCP level %q", s)
}

func parseRendition(items chan item) (*Rendition, error) {
	var rend Rendition
	var err error
	for it := range items {
		if it.typ != itemAttrName {
			return nil, fmt.Errorf("expected attribute name, got %s", it)
		}
		attr := it
		it = <-items
		if it.typ != itemEquals {
			return nil, fmt.Errorf("parse %s: expected =, got %s", attr, it)
		}
		it = <-items
		switch attr.val {
		case "TYPE":
			rend.Type, err = parseMediaType(it.val)
			if err != nil {
				return nil, fmt.Errorf("parse media type: %w", err)
			}
		case "URI":
			rend.URI = strings.Trim(it.val, `"`)
		case "GROUP-ID":
			rend.Group = strings.Trim(it.val, `"`)
		case "LANGUAGE":
			rend.Language = strings.Trim(it.val, `"`)
		case "ASSOC-LANGUAGE":
			rend.AssocLanguage = strings.Trim(it.val, `"`)
		case "NAME":
			rend.Name = strings.Trim(it.val, `"`)
		case "DEFAULT", "AUTOSELECT", "FORCED":
			b, err := parseBool(it.val)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", attr, err)
			}
			if attr.val == "DEFAULT" {
				rend.Default = b
			} else if attr.val == "AUTOSELECT" {
				rend.AutoSelect = b
			} else if attr.val == "FORCED" {
				rend.Forced = b
			}
		case "INSTREAM-ID":
			rend.InstreamID, err = parseCCInfo(it.val)
			if err != nil {
				return nil, fmt.Errorf("parse instream-id: %w", err)
			}
		case "CHARACTERISTICS":
			rend.Characteristics = strings.Split(it.val, ",")
		case "CHANNELS":
			rend.Channels = strings.Split(strings.Trim(it.val, `"`), "/")
		default:
			return nil, fmt.Errorf("unknown rendition attribute %s", attr.val)
		}
		it = <-items
		switch it.typ {
		case itemError:
			return nil, fmt.Errorf("next attribute: %s", it.val)
		case itemComma:
			continue
		case itemNewline:
			return &rend, nil
		default:
			return nil, fmt.Errorf("next attribute: expected comma or newline, got %s", it)
		}
	}
	return &rend, nil
}

func parseMediaType(s string) (MediaType, error) {
	for t := MediaAudio; t <= MediaClosedCaptions; t++ {
		if t.String() == s {
			return t, nil
		}
	}
	return 0, fmt.Errorf("unknown media type %s", s)
}

func parseBool(s string) (bool, error) {
	if s == "YES" {
		return true, nil
	} else if s == "NO" {
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean string %s", s)
}

// parseCCInfo parses a CCInfo attribute from s as specified in RFC 8216 section 4.4.6.1.
func parseCCInfo(s string) (*CCInfo, error) {
	// shortest possible is 3 chars, "CC0", "CC1" etc.
	if len(s) < 3 {
		return nil, fmt.Errorf("too short")
	}
	if s[:2] == "CC" {
		// MUST have one of the values: "CC1", "CC2", "CC3", "CC4"
		switch {
		case len(s) == 3 && s[2] >= '1' && s[2] <= '4':
			i := int(s[2] - '0')
			return &CCInfo{i, false}, nil
		default:
			return nil, fmt.Errorf("invalid closed caption %s", s)
		}
	}
	// SERVICE00
	if len(s) < 8 {
		return nil, fmt.Errorf("invalid keyword %s", s)
	}
	if s[:6] != "SERVICE" {
		return nil, fmt.Errorf("expected keyword %q, got %q", "SERVICE", s[:6])
	} else if len(s) > 9 {
		return nil, fmt.Errorf("service too long")
	}
	i, err := strconv.Atoi(s[6:])
	if err != nil {
		return nil, fmt.Errorf("parse service block number: %w", err)
	}
	if i < 1 || i > 63 {
		return nil, fmt.Errorf("invalid service block number %d", i)
	}
	return &CCInfo{i, true}, nil
}

func parsePlaylistType(it item) (PlaylistType, error) {
	if it.typ != itemAttrName {
		return 0, fmt.Errorf("got %s, want item type %s", it, itemString)
	}
	switch it.val {
	case "EVENT":
		return PlaylistEvent, nil
	case "VOD":
		return PlaylistVOD, nil
	}
	return 0, fmt.Errorf("illegal playlist type %q", it.val)
}

func parseTargetDuration(it item) (time.Duration, error) {
	if it.typ != itemAttrName && it.typ != itemNumber {
		return 0, fmt.Errorf("got %s: want attribute name or number", it)
	}
	i, err := strconv.Atoi(it.val)
	if err != nil {
		return 0, err
	}
	return time.Duration(i) * time.Second, nil
}

func parseByteRange(s string) (ByteRange, error) {
	offset, until, found := strings.Cut(s, "@")
	if !found {
		n, err := strconv.Atoi(offset)
		if err != nil {
			return ByteRange{}, err
		}
		return ByteRange{n, 0}, nil
	}
	n, err := strconv.Atoi(offset)
	if err != nil {
		return ByteRange{}, err
	}
	nn, err := strconv.Atoi(until)
	if err != nil {
		return ByteRange{}, err
	}
	return ByteRange{n, nn}, nil
}
