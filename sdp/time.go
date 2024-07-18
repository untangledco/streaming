package sdp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// number of seconds from the zero time used in SDP; 1900-01-01T00:00Z
// to the Unix epoch.
const sinceTimeZero = 2208988800

func parseTimes(s string) ([2]time.Time, error) {
	var times [2]time.Time
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return times, fmt.Errorf("bad number of fields %d: need 2", len(fields))
	}
	start, err := strconv.Atoi(fields[0])
	if err != nil {
		return times, fmt.Errorf("parse start time: %w", err)
	}
	if start != 0 {
		times[0] = time.Unix(int64(start-sinceTimeZero), 0).UTC()
	}
	end, err := strconv.Atoi(fields[1])
	if err != nil {
		return times, fmt.Errorf("parse end time: %w", err)
	}
	if end != 0 {
		times[1] = time.Unix(int64(end-sinceTimeZero), 0).UTC()
	}
	return times, nil
}

// Repeat represents a session's repetition cycle as described in
// RFC 8866 section 5.10.
type Repeat struct {
	Interval time.Duration   // duration between each repetition cycle
	Active   time.Duration   // planned duration of each session
	Offsets  []time.Duration // duration(s) between each session
}

func (rp *Repeat) String() string {
	// TODO(otl): print with no decimal places?
	s := fmt.Sprintf("r=%f %f ", rp.Interval.Round(time.Second).Seconds(), rp.Active.Round(time.Second).Seconds())
	if len(rp.Offsets) > 0 {
		ss := make([]string, len(rp.Offsets))
		for i := range rp.Offsets {
			ss[i] = fmt.Sprintf("%f", rp.Offsets[i].Round(time.Second).Seconds())
		}
		s += strings.Join(ss, " ")
	}
	return strings.TrimSpace(s)
}

func parseRepeat(s string) (Repeat, error) {
	// guard against negative durations, decimals.
	// these are valid for time.ParseDuration, but not for our Repeat.
	if strings.Contains(s, "-") || strings.Contains(s, ".") {
		return Repeat{}, errors.New("invalid duration")
	}

	fields := strings.Fields(s)
	if len(fields) < 3 {
		return Repeat{}, fmt.Errorf("short line: have %d, want at least %d fields", len(fields), 3)
	}

	var repeat Repeat
	var err error
	repeat.Interval, err = parseDuration(fields[0])
	if err != nil {
		return Repeat{}, fmt.Errorf("parse interval %s: %w", fields[0], err)
	}
	repeat.Active, err = parseDuration(fields[1])
	if err != nil {
		return Repeat{}, fmt.Errorf("parse active duration %s: %w", fields[1], err)
	}
	for _, s := range fields[2:] {
		offset, err := parseDuration(s)
		if err != nil {
			return Repeat{}, fmt.Errorf("parse offset %s: %w", s, err)
		}
		repeat.Offsets = append(repeat.Offsets, offset)
	}
	return repeat, nil
}

func parseDuration(s string) (time.Duration, error) {
	// a bare int, like 86400
	i, err := strconv.Atoi(s)
	if err == nil {
		return time.Duration(i) * time.Second, nil
	}

	// a duration string like 24h
	dur, err := time.ParseDuration(s)
	if err == nil {
		return dur, nil
	}

	// a duration string with days suffix, like 1d
	// [0-9]+d
	if !strings.HasSuffix(s, "d") {
		return 0, fmt.Errorf("bad duration: expected d suffix for days")
	}
	j, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return 0, fmt.Errorf("parse days: %w", err)
	}
	return time.Duration(j) * 24 * time.Hour, nil
}

type TimeAdjustment struct {
	When   time.Time
	Offset time.Duration
}

func (t TimeAdjustment) String() string {
	return fmt.Sprintf("%d %f", t.When.Unix()+sinceTimeZero, t.Offset.Round(time.Second).Seconds())
}

func parseAdjustments(line string) ([]TimeAdjustment, error) {
	fields := strings.Fields(line)
	if len(fields)%2 != 0 {
		return nil, fmt.Errorf("odd field count %d", len(fields))
	}
	var adjustments []TimeAdjustment
	for i := 0; i < len(fields); i += 2 {
		var adj TimeAdjustment
		t, err := strconv.Atoi(fields[i])
		if err != nil {
			return nil, fmt.Errorf("time %s: %w", fields[i], err)
		}
		adj.When = time.Unix(int64(t-sinceTimeZero), 0).UTC()
		adj.Offset, err = parseDuration(fields[i+1])
		if err != nil {
			return nil, fmt.Errorf("offset %s: %w", fields[i+1], err)
		}
		adjustments = append(adjustments, adj)
	}
	return adjustments, nil
}

// Now returns the current SDP timestamp; the number of seconds
// since 1900-01-01T00:00Z.
func Now() int64 {
	return time.Now().Unix() + sinceTimeZero
}
