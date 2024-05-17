package m3u8

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Media segment tags specified in RFC 8216 section 4.4.4.
const (
	tagSegmentDuration = "#EXTINF"
	tagByteRange       = "#EXT-X-BYTERANGE"
	tagDiscontinuity   = "#EXT-X-DISCONTINUITY"
	tagKey             = "#EXT-X-KEY"
	tagMap             = "#EXT-X-MAP"
	tagDateTime        = "#EXT-X-PROGRAM-DATE-TIME"
	tagGap             = "#EXT-X-GAP"
	tagBitrate         = "#EXT-X-BITRATE"
	tagPart            = "#EXT-X-PART"
)

// parseSegment returns the next segment from items and the leading
// item which indecated the start of a segment.
func parseSegment(items chan item, leading item) (*Segment, error) {
	var seg Segment
	switch leading.typ {
	case itemTag:
		switch leading.val {
		case tagSegmentDuration:
			it := <-items
			dur, err := parseSegmentDuration(it)
			if err != nil {
				return nil, fmt.Errorf("parse segment duration: %w", err)
			}
			seg.Duration = dur
		}
	}
	for it := range items {
		if it.typ == itemError {
			return nil, errors.New(it.val)
		}
		switch it.typ {
		case itemURL:
			seg.URI = it.val
			return &seg, nil
		case itemTag:
			switch it.val {
			case tagSegmentDuration:
				it = <-items
				dur, err := parseSegmentDuration(it)
				if err != nil {
					return nil, fmt.Errorf("parse segment duration: %w", err)
				}
				seg.Duration = dur
			case tagByteRange:
				it = <-items
				if it.typ != itemString {
					return nil, fmt.Errorf("parse byte range: got %s, want item type string", it)
				}
				r, err := parseByteRange(it.val)
				if err != nil {
					return nil, fmt.Errorf("parse byte range: %w", err)
				}
				seg.Range = r
			case tagDiscontinuity:
				seg.Discontinuity = true
			case tagKey:
				return nil, fmt.Errorf("parsing %s unsupported", it)
			default:
				return nil, fmt.Errorf("parsing %s unsupported", it)
			}
		}
	}
	return nil, fmt.Errorf("no url")
}

func parseSegmentDuration(it item) (time.Duration, error) {
	if it.typ != itemAttrName && it.typ != itemNumber {
		return 0, fmt.Errorf("got %s: want attribute name or number", it)
	}
	// Some numbers can be converted straight to ints, e.g.:
	// 	10
	// 	10.000
	// Others need to be converted from floating point, e.g:
	// 	9.967
	// Try the easiest paths first.
	if !strings.Contains(it.val, ".") {
		i, err := strconv.Atoi(it.val)
		if err != nil {
			return 0, err
		}
		return time.Duration(i) * time.Second, nil
	}
	// 10.000
	before, after, _ := strings.Cut(it.val, ".")
	var allZeroes = true
	for r := range after {
		if r != '0' {
			allZeroes = false
		}
	}
	if allZeroes {
		i, err := strconv.Atoi(before)
		if err != nil {
			return 0, err
		}
		return time.Duration(i) * time.Second, nil
	}
	seconds, err := strconv.ParseFloat(it.val, 32)
	if err != nil {
		return 0, err
	}
	// precision based on a 90KHz clock.
	microseconds := seconds * 1e6
	return time.Duration(microseconds) * time.Microsecond, nil
}
