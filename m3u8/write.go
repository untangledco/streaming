package m3u8

import (
	"fmt"
	"io"
	"time"
)

func Encode(w io.Writer, p *Playlist) error {
	fmt.Fprintln(w, "#EXTM3U")
	fmt.Fprintf(w, "%s:%d\n", tagVersion, p.Version)
	fmt.Fprintf(w, "%s:%s\n", tagPlaylistType, p.Type)
	fmt.Fprintf(w, "%s:%d\n", tagTargetDuration, p.TargetDuration/time.Second)
	for _, seg := range p.Segments {
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
