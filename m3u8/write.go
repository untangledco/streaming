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
	fmt.Fprintf(w, "%s:%d\n", tagVersion, p.Version)
	fmt.Fprintf(w, "%s:%s\n", tagPlaylistType, p.Type)
	fmt.Fprintf(w, "%s:%d\n", tagTargetDuration, p.TargetDuration/time.Second)
	for _, seg := range p.Segments {
		if seg.Discontinuity {
			fmt.Fprintln(w, tagDiscontinuity)
		}
		if seg.DateRange != nil {
			if err := writeDateRange(w, seg.DateRange); err != nil {
				return fmt.Errorf("write date range: %w", err)
			}
		}
		us := seg.Duration / time.Microsecond
		// we do .03f for the same precision as test-streams.mux.dev.
		fmt.Fprintf(w, "%s:%.03f\n", tagSegmentDuration, float32(us)/1e6)
		fmt.Fprintln(w, seg.URI)
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
