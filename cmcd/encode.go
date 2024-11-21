package cmcd

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

func (r Request) Encode() string {
	var attrs []string
	if r.BufLength > 0 {
		a := fmt.Sprintf("bl=%d", r.BufLength.Round(100*time.Millisecond).Milliseconds())
		attrs = append(attrs, a)
	}
	if r.Deadline > 0 {
		a := fmt.Sprintf("dl=%d", r.Deadline.Round(100*time.Millisecond).Milliseconds())
		attrs = append(attrs, a)
	}
	if r.Throughput > 0 {
		attrs = append(attrs, fmt.Sprintf("mtp=%d", r.Throughput))
	}
	if r.Next != "" {
		a := fmt.Sprintf("nor=%q", url.QueryEscape(r.Next))
		attrs = append(attrs, a)
	}
	if r.NextRange != [2]int{0, 0} {
		a := fmt.Sprintf("nrr=%q", r.NextRange)
		attrs = append(attrs, a)
	}
	if r.Startup == true {
		attrs = append(attrs, "su")
	}
	return strings.Join(attrs, ",")
}

func (o Object) Encode() string {
	var attrs []string
	if o.Bitrate > 0 {
		attrs = append(attrs, fmt.Sprintf("br=%d", o.Bitrate))
	}
	if o.Duration > 0 {
		a := fmt.Sprintf("d=%d", o.Duration.Milliseconds())
		attrs = append(attrs, a)
	}
	if o.Type != "" {
		// not a quoted string as these are reserved keywords.
		attrs = append(attrs, fmt.Sprintf("ot=%s", o.Type))
	}
	if o.TopBitrate > 0 {
		attrs = append(attrs, fmt.Sprintf("tb=%d", o.TopBitrate))
	}
	return strings.Join(attrs, ",")
}

func (s Status) Encode() string {
	var attrs []string
	if s.Starved {
		attrs = append(attrs, "bs")
	}
	if s.MaxThroughput > 0 {
		attrs = append(attrs, fmt.Sprintf("rtp=%d", s.MaxThroughput))
	}
	return strings.Join(attrs, ",")
}

func (s Session) Encode() string {
	var attrs []string
	if s.ID != "" {
		attrs = append(attrs, fmt.Sprintf("sid=%q", s.ID))
	}
	if s.ContentID != "" {
		attrs = append(attrs, fmt.Sprintf("cid=%q", s.ContentID))
	}
	// "SHOULD only be sent if not equal to 1": CTA-5004 page 10.
	if s.PlayRate != RealTime {
		attrs = append(attrs, fmt.Sprintf("pr=%v", s.PlayRate))
	}
	switch s.Format {
	case FormatDASH, FormatHLS, FormatSmooth, FormatOther:
		attrs = append(attrs, fmt.Sprintf("sf=%s", s.Format))
	}
	return strings.Join(attrs, ",")
}
