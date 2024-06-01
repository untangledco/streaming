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
		if r.Name == "" {
			return fmt.Errorf("empty name")
		}
		fmt.Fprintf(w, tagRendition+":")
		// TODO(otl): use string slice, then strings.Join(s, ",")
		// then we don't worry about errors from w.Write
		fmt.Fprintf(w, "NAME=%q,", r.Name)
		rname := fmt.Sprintf("rendition %s", r.Name)
		if r.Type > MediaClosedCaptions {
			return fmt.Errorf("%s: unknown type %s", rname, r.Type)
		}
		fmt.Fprintf(w, "TYPE=%s,", r.Type)
		if r.URI != "" {
			fmt.Fprintf(w, "URI=%q,", r.URI)
		}
		if r.Group == "" {
			return fmt.Errorf("%s: empty group", rname)
		}
		fmt.Fprintf(w, "GROUP-ID=%q,", r.Group)
		if r.Language != "" {
			fmt.Fprintf(w, "LANGUAGE=%q,", r.Language)
		}
		if r.AssocLanguage != "" {
			fmt.Fprintf(w, "ASSOC-LANGUAGE=%q,", r.AssocLanguage)
		}
		if r.Default {
			fmt.Fprint(w, "DEFAULT=YES,")
		}
		if r.AutoSelect {
			fmt.Fprint(w, "AUTOSELECT=YES,")
		}
		if r.Forced {
			fmt.Fprint(w, "FORCED=YES,")
		}

		if r.Type != MediaClosedCaptions && r.InstreamID != nil {
			return fmt.Errorf("%s: instream-id set but type is %s", rname, r.Type)
		}
		if r.Type == MediaClosedCaptions {
			if r.InstreamID == nil {
				return fmt.Errorf("%s: nil instream-id", rname)
			}
			fmt.Fprintf(w, "INSTREAM-ID=%q,", r.InstreamID)
		}
		if len(r.Characteristics) > 0 {
			fmt.Fprintf(w, "CHARACTERISTICS=%q,", strings.Join(r.Characteristics, ","))
		}
		if len(r.Channels) > 0 {
			fmt.Fprintf(w, "CHANNELS=%q,", strings.Join(r.Channels, "/"))
		}
		fmt.Fprintln(w)
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
	fmt.Fprint(w, tagDateRange+":")
	var attrs []string
	if dr.ID != "" {
		attrs = append(attrs, fmt.Sprintf("ID=%q", dr.ID))
	}
	if dr.Class != "" {
		attrs = append(attrs, fmt.Sprintf("CLASS=%q", dr.Class))
	}
	if !dr.Start.IsZero() {
		attrs = append(attrs, fmt.Sprintf("START-DATE=%q", dr.Start.Format(time.RFC3339)))
	}
	if !dr.End.IsZero() {
		attrs = append(attrs, fmt.Sprintf("END-DATE=%q", dr.End.Format(time.RFC3339)))
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
		attrs = append(attrs, "END-ON-NEXT:YES")
	}
	if _, err := fmt.Fprintln(w, strings.Join(attrs, ",")); err != nil {
		return err
	}
	return nil
}

func writeMap(w io.Writer, m Map) (n int, err error) {
	if m.ByteRange != [2]int{0, 0} {
		return fmt.Fprintf(w, "%s:URI=%q,BYTERANGE=%s\n", tagMap, m.URI, m.ByteRange)
	}
	return fmt.Fprintf(w, "%s:URI=%q\n", tagMap, m.URI)
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
