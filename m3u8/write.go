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

	for _, v := range p.Variants {
		fmt.Fprint(w, tagVariant+":")
		if v.Bandwidth > 0 {
			fmt.Fprintf(w, "BANDWIDTH=%d,", v.Bandwidth)
		}
		if v.AverageBandwidth > 0 {
			fmt.Fprintf(w, "AVERAGE-BANDWIDTH=%d,", v.AverageBandwidth)
		}
		if len(v.Codecs) > 0 {
			fmt.Fprintf(w, "CODECS=%q,", strings.Join(v.Codecs, ","))
		}
		if v.Resolution != [2]int{0, 0} {
			fmt.Fprintf(w, "RESOLUTION=%dx%d,", v.Resolution[0], v.Resolution[1])
		}
		if v.FrameRate > 0 {
			fmt.Fprintf(w, "FRAME-RATE=%f,", v.FrameRate)
		}
		if v.HDCP != HDCPNone {
			fmt.Fprintf(w, "HDCP-LEVEL=%s,", v.HDCP)
		}
		if v.Audio != "" {
			fmt.Fprintf(w, "AUDIO=%q,", v.Audio)
		}
		if v.Video != "" {
			fmt.Fprintf(w, "VIDEO=%q,", v.Video)
		}
		if v.Subtitles != "" {
			fmt.Fprintf(w, "SUBTITLES=%q,", v.Subtitles)
		}
		if v.ClosedCaptions != "" && v.ClosedCaptions != NoClosedCaptions {
			fmt.Fprintf(w, "CLOSED-CAPTIONS=%q,", v.ClosedCaptions)
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, v.URI)
	}

	for _, sd := range p.SessionData {
		fmt.Fprintf(w, tagSessionData+":")
		fmt.Fprintf(w, "DATA-ID=%q,", sd.ID)
		if sd.URI != "" && sd.Value != "" {
			return fmt.Errorf("must contain either a VALUE(%s) or URI(%s) attribute", sd.Value, sd.URI)
		}
		if sd.Value != "" {
			fmt.Fprintf(w, "VALUE=%q", sd.Value)
		}
		if sd.URI != "" {
			fmt.Fprintf(w, "URI=%q", sd.URI)
		}
		if sd.Language != "" {
			fmt.Fprintf(w, ",LANGUAGE=%q", sd.Language)
		}
		fmt.Fprintln(w)
	}

	if p.End {
		fmt.Fprintln(w, tagEndList)
	}
	return nil
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
		b, err := scte35.EncodeSpliceInfo(dr.CueIn)
		if err != nil {
			return fmt.Errorf("encode cue in: %w", err)
		}
		attrs = append(attrs, fmt.Sprintf("SCTE35-IN=0x%s", hex.EncodeToString(b)))
	}
	if dr.CueOut != nil {
		b, err := scte35.EncodeSpliceInfo(dr.CueOut)
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
