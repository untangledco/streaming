package m3u8

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/untangledco/streaming/scte35"
)

func Encode(w io.Writer, p *Playlist) error {
	fmt.Fprintln(w, "#EXTM3U")
	if p.Version > 0 {
		fmt.Fprintf(w, "%s:%d\n", tagVersion, p.Version)
	}
	if p.Type != PlaylistNone {
		fmt.Fprintf(w, "%s:%s\n", tagPlaylistType, p.Type)
	}
	if p.IndependentSegments {
		fmt.Fprintln(w, tagIndependentSegments)
	}
	if p.TargetDuration > 0 {
		fmt.Fprintf(w, "%s:%d\n", tagTargetDuration, p.TargetDuration/time.Second)
	}
	fmt.Fprintf(w, "%s:%d\n", tagMediaSequence, p.Sequence)

	if _, err := writeSegments(w, p.Segments); err != nil {
		return fmt.Errorf("write segments: %w", err)
	}

	for _, r := range p.Media {
		if _, err := writeRendition(w, r); err != nil {
			return fmt.Errorf("rendition %s: %w", r.Name, err)
		}
	}

	for i, v := range p.Variants {
		if _, err := writeVariant(w, &v); err != nil {
			return fmt.Errorf("write variant %d: %w", i, err)
		}
	}

	for i, sd := range p.SessionData {
		if _, err := writeSessionData(w, sd); err != nil {
			return fmt.Errorf("write session data %d: %w", i, err)
		}
	}

	if p.End {
		fmt.Fprintln(w, tagEndList)
	}
	return nil
}

func writeVariant(w io.Writer, v *Variant) (n int, err error) {
	if v.Bandwidth <= 0 {
		return 0, fmt.Errorf("invalid bandwidth %d: must be larger than zero", v.Bandwidth)
	}
	if v.URI == "" {
		return 0, fmt.Errorf("empty URI")
	}
	return fmt.Fprintln(w, v)
}

func writeDateRange(w io.Writer, dr *DateRange) error {
	if dr.ID == "" {
		return fmt.Errorf("empty ID")
	} else if dr.Start.IsZero() {
		return fmt.Errorf("zero start time")
	}

	var attrs []string
	attrs = append(attrs, fmt.Sprintf("ID=%q", dr.ID))
	attrs = append(attrs, fmt.Sprintf("START-DATE=%q", dr.Start.Format(rfc3339Milli)))
	if !dr.End.IsZero() {
		attrs = append(attrs, fmt.Sprintf("END-DATE=%q", dr.End.Format(rfc3339Milli)))
	}
	if dr.Class != "" {
		attrs = append(attrs, fmt.Sprintf("CLASS=%q", dr.Class))
	}
	// TODO(otl): dr.Duration, dr.Planned. Differentiate zero value and user-set zero.
	// TODO(otl): dr.Custom.
	// TODO(otl): dr.CueCommand, when to write this versuse cuein, cueout.
	if dr.CueIn != nil {
		b, err := scte35.Encode(dr.CueIn)
		if err != nil {
			return fmt.Errorf("encode cue in: %w", err)
		}
		attrs = append(attrs, fmt.Sprintf("SCTE35-IN=0x%s", hex.EncodeToString(b)))
	}
	if dr.CueOut != nil {
		b, err := scte35.Encode(dr.CueOut)
		if err != nil {
			return fmt.Errorf("encode cue out: %w", err)
		}
		attrs = append(attrs, fmt.Sprintf("SCTE35-OUT=0x%s", hex.EncodeToString(b)))
	}
	if dr.EndOnNext {
		if dr.Class == "" {
			return fmt.Errorf("empty class with end-on-next set")
		} else if !dr.End.IsZero() {
			return fmt.Errorf("non-zero end time with end-on-next set")
		} else if dr.Duration > 0 {
			return fmt.Errorf("non-zero duration %s with end-on-next set", dr.Duration)
		}
		attrs = append(attrs, "END-ON-NEXT:YES")
	}
	tag := tagDateRange + ":" + strings.Join(attrs, ",")
	_, err := fmt.Fprintln(w, tag)
	return err
}

func writeRendition(w io.Writer, r Rendition) (n int, err error) {
	if r.Name == "" {
		return 0, fmt.Errorf("empty name")
	} else if r.Group == "" {
		return 0, fmt.Errorf("empty group")
	}

	if r.Type > MediaClosedCaptions {
		return 0, fmt.Errorf("unknown type %s", r.Type)
	}
	if r.Type != MediaClosedCaptions && r.InstreamID != nil {
		return 0, fmt.Errorf("instream-id set but type is %s", r.Type)
	} else if r.Type == MediaClosedCaptions && r.InstreamID == nil {
		return 0, fmt.Errorf("nil instream-id")
	}
	return fmt.Fprintln(w, r)
}

func writeSessionData(w io.Writer, sd SessionData) (n int, err error) {
	if sd.ID == "" {
		return 0, fmt.Errorf("ID not set")
	}
	if sd.URI != "" && sd.Value != "" {
		return 0, fmt.Errorf("only one of Value or URI may be set")
	}
	return fmt.Fprintln(w, sd)
}
