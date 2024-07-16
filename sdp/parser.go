package sdp

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type parser struct {
	*bufio.Scanner
	err error
	// Field name and value from the current line.
	// TODO(otl): rename? key, value is very non-specific...
	key, value string
	next       []string // expected next field names

	session Session
}

var ftab = [...]string{"i", "u", "e", "p", "c", "b", "t", "r", "z", "a", "m", "a"}

var mtab = [...]string{"i", "c", "b", "a", "m"}

func (p *parser) scan() bool {
	if !p.Scan() {
		p.err = p.Err()
		return false
	}
	line := strings.TrimSpace(p.Text())
	if line == "" {
		p.err = fmt.Errorf("illegal empty line")
		return false
	}
	k, v, found := strings.Cut(line, "=")
	if !found {
		p.err = fmt.Errorf("parse field %q, missing %q", k, "=")
		return false
	}
	p.key = k
	p.value = v

	if p.next != nil {
		for i := range p.next {
			if p.next[i] == p.key {
				return true
			}
		}
		p.err = fmt.Errorf("unexpected field %q: expected one of %q", k, p.next)
		return false
	}
	return true
}

func (p *parser) parse() error {
	next := "v"
	for p.scan() {
		if p.key != next {
			return fmt.Errorf("expected key %q, found %q", next, p.key)
		}
		switch p.key {
		case "v":
			i, err := strconv.Atoi(p.value)
			if err != nil {
				return fmt.Errorf("parse version: %w", err)
			}
			if i != 0 {
				return fmt.Errorf("unsupported version %d", i)
			}
			next = "o"
		case "o":
			o, err := parseOrigin(p.value)
			if err != nil {
				return fmt.Errorf("parse origin: %w", err)
			}
			p.session.Origin = o
			next = "s"
		case "s":
			if p.value == "" {
				return fmt.Errorf("empty name")
			}
			p.session.Name = p.value
			return p.parseOptional()
		}
	}
	return p.err
}

func (p *parser) parseOptional() error {
	p.next = ftab[:]
	for p.scan() {
		switch p.key {
		case "i":
			p.session.Info = p.value
			p.next = ftab[1:]
		case "u":
			u, err := url.Parse(p.value)
			if err != nil {
				return fmt.Errorf("parse uri: %w", err)
			}
			p.session.URI = u
			p.next = ftab[2:]
		case "e":
			addr, err := parseEmail(p.value)
			if err != nil {
				return fmt.Errorf("parse email: %w", err)
			}
			p.session.Email = addr
			p.next = ftab[3:]
		case "p":
			p.session.Phone = cleanPhone(p.value)
			p.next = ftab[4:]
		case "c":
			conn, err := parseConnInfo(p.value)
			if err != nil {
				return fmt.Errorf("parse connection info: %w", err)
			}
			p.session.Connection = &conn
			p.next = ftab[5:]
		case "b":
			bw, err := parseBandwidth(p.value)
			if err != nil {
				return fmt.Errorf("parse bandwidth line %q: %w", p.value, err)
			}
			p.session.Bandwidth = &bw
			p.next = ftab[6:]
		case "t":
			when, err := parseTimes(p.value)
			if err != nil {
				return fmt.Errorf("parse time description: %w", err)
			}
			p.session.Time = when
			p.next = ftab[7:]
		case "r":
			repeat, err := parseRepeat(p.value)
			if err != nil {
				return fmt.Errorf("parse repeat: %w", err)
			}
			p.session.Repeat = &repeat
			p.next = ftab[8:]
		case "z":
			return fmt.Errorf("parse time desc %s not yet implemented", p.value)
		case "a":
			p.session.Attributes = strings.Fields(p.value)
			p.next = ftab[9:]
		case "m":
			m, err := parseMedia(p.value)
			if err != nil {
				return fmt.Errorf("parse media info from %q: %w", p.value, err)
			}
			p.session.Media = append(p.session.Media, m)
			p.next = mtab[:]
			return p.parseMedia()
		}
	}
	return p.err
}

func (p *parser) parseMedia() error {
	var media *Media
	if len(p.session.Media) > 0 {
		media = &p.session.Media[len(p.session.Media)-1]
	}
	for p.scan() {
		switch p.key {
		case "i":
			media.Title = p.value
			p.next = mtab[1:]
		case "c":
			conn, err := parseConnInfo(p.value)
			if err != nil {
				return fmt.Errorf("parse connection info: %w", err)
			}
			media.Connection = &conn
			p.next = mtab[2:]
		case "b":
			bw, err := parseBandwidth(p.value)
			if err != nil {
				return fmt.Errorf("parse bandwidth: %w", err)
			}
			media.Bandwidth = &bw
			p.next = mtab[3:]
		case "a":
			media.Attributes = strings.Fields(p.value)
			p.next = mtab[4:]
		case "m":
			m, err := parseMedia(p.value)
			if err != nil {
				return fmt.Errorf("parse media description: %w", err)
			}
			p.session.Media = append(p.session.Media, m)
			media = &p.session.Media[len(p.session.Media)-1]
			p.next = mtab[:]
		default:
			return fmt.Errorf("unsupported field char %s", p.key)
		}
	}
	return p.err
}

// number of seconds since the zero time used in SDP 1900-01-01T00:00Z.
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

type Repeat struct {
	Interval time.Duration
	Active   time.Duration
	Offsets  []time.Duration
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
