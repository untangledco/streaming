package scte35

import (
	"encoding/binary"
	"fmt"
	"time"
)

// GPS epoch is 1980-01-06T00:00:00Z
var gpsEpoch time.Time = time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)

// Command represents a splice command described in
// SCTE 35 section 9.7.
type Command struct {
	Type     CommandType
	Schedule []Event // SpliceSchedule
	Insert   *Insert
	// Number of ticks of a 90KHz clock since midnight UTC.
	// TODO(otl): use time.Time here instead,
	// then calculate ticks when converting to wire format?
	TimeSignal *uint64
	Private    *PrivateCommand
}

type CommandType uint8

const (
	SpliceNull           CommandType = 0
	SpliceSchedule                   = 0x04
	SpliceInsert                     = 0x05
	TimeSignal                       = 0x06
	BandwidthReservation             = 0x07
	Private                          = 0xff
)

func (t CommandType) String() string {
	switch t {
	case SpliceNull:
		return "splice_null"
	case SpliceSchedule:
		return "splice_schedule"
	case SpliceInsert:
		return "splice_insert"
	case TimeSignal:
		return "time_signal"
	case BandwidthReservation:
		return "bandwidth_reservation"
	case Private:
		return "private_command"
	}
	return "reserved"
}

func encodeCommand(c *Command) ([]byte, error) {
	switch c.Type {
	case SpliceNull, BandwidthReservation:
		// Since SpliceNull == 0 (default) check if we've
		// accidentally set another field.
		if c.Schedule != nil {
			return nil, fmt.Errorf("command %s has non-nil schedule", c.Type)
		} else if c.Insert != nil {
			return nil, fmt.Errorf("command %s has non-nil Insert", c.Type)
		} else if c.TimeSignal != nil {
			return nil, fmt.Errorf("command %s has non-nil TimeSignal", c.Type)
		} else if c.Private != nil {
			return nil, fmt.Errorf("command %s has non-nil Private", c.Type)
		}
		return nil, nil
	case SpliceSchedule:
		b, err := packEvents(c.Schedule)
		if err != nil {
			return b, fmt.Errorf("pack events: %w", err)
		}
		return b, nil
	case SpliceInsert:
		return encodeInsert(c.Insert), nil
	case TimeSignal:
		if c.TimeSignal == nil {
			return nil, fmt.Errorf("cannot encode nil TimeSignal")
		}
		b := encodeSpliceTime(*c.TimeSignal)
		return b[:], nil
	case Private:
		return encodePrivateCommand(c.Private), nil
	default:
		return nil, fmt.Errorf("encoding command %s unsupported", c.Type)
	}
}

// Event is a single event within a splice_schedule.
type Event struct {
	ID uint32
	// Indicates a previously sent event identified by ID should
	// be cancelled.
	Cancel bool

	OutOfNetwork bool
	// TODO(otl): should always be true? should we support
	// deprecated Component Splice Mode?
	// see section 9.7.2.1.
	// ProgramSplice bool
	SpliceTime    time.Time
	BreakDuration *BreakDuration

	ProgramID     uint16
	AvailNum      uint8
	AvailExpected uint8
	// Indicates the event's ID is prepared in the method
	// described in SCTE 35 section 9.3.3.
	// TODO(otl): can we calculate this at runtime?
	// See https://github.com/untangledco/streaming/issues/2
	idCompliance bool
}

func packEvents(events []Event) ([]byte, error) {
	if len(events) > 255 {
		return nil, fmt.Errorf("too many events (%d), need 255 or less", len(events))
	}
	var packed []byte
	packed[0] = uint8(len(events))
	for i := range events {
		b := packEvent(&events[i])
		packed = append(packed, b...)
	}
	return packed, nil
}

func packEvent(e *Event) []byte {
	// length is e.ID + flags
	p := make([]byte, 4+1)
	binary.BigEndian.PutUint32(p[:4], e.ID)
	if e.Cancel {
		p[4] |= 1 << 7
	}
	if e.idCompliance {
		p[4] |= 1 << 6
	}
	// 6 remaining bits are reserved.

	if !e.Cancel {
		p = append(p, 0x00)
		if e.OutOfNetwork {
			p[5] |= 1 << 7
		}
		// assume program_splice is always set;
		// we don't support component splice mode.
		p[5] |= 1 << 6
		if e.BreakDuration != nil {
			p[5] |= 1 << 5
		}
		// 5 remaining bits are reserved

		seconds := e.SpliceTime.Sub(gpsEpoch) / time.Second
		p = binary.BigEndian.AppendUint32(p, uint32(seconds))

		if e.BreakDuration != nil {
			bd := packBreakDuration(e.BreakDuration)
			p = append(p, bd[:]...)
		}
	}

	p = binary.BigEndian.AppendUint16(p, e.ProgramID)
	p = append(p, e.AvailNum)
	p = append(p, e.AvailExpected)
	return p
}

type PrivateCommand struct {
	ID   uint32
	Data []byte
}

func encodePrivateCommand(c *PrivateCommand) []byte {
	buf := make([]byte, 4+len(c.Data))
	binary.BigEndian.PutUint32(buf[:4], c.ID)
	copy(buf[4:], c.Data)
	return buf
}

func decodePrivateCommand(b []byte) (PrivateCommand, error) {
	if len(b) < 4 {
		return PrivateCommand{}, fmt.Errorf("need at least 4 bytes, have %d", len(b))
	}
	return PrivateCommand{
		ID:   binary.BigEndian.Uint32(b[:4]),
		Data: b[4:],
	}, nil
}

// Insert represents the splice_insert command
// as specified in SCTE 35 section 9.7.3.
type Insert struct {
	ID           uint32
	Cancel       bool
	OutOfNetwork bool
	Immediate    bool
	// Number of ticks of a 90KHz clock.
	SpliceTime    *uint64
	Duration      *BreakDuration
	ProgramID     uint16
	AvailNum      uint8
	AvailExpected uint8
	// Indicates the event's ID is prepared in the method
	// described in SCTE 35 section 9.3.3.
	// TODO(otl): can we calculate this at runtime?
	// See https://github.com/untangledco/streaming/issues/2
	idCompliance bool
}

func encodeInsert(ins *Insert) []byte {
	buf := make([]byte, 4+1) // uint32 + 1 byte
	binary.BigEndian.PutUint32(buf[:4], ins.ID)
	buf[4] |= 0x7f // toggle reserved bits
	if ins.Cancel {
		buf[4] |= (1 << 7) // toggle unreserved bit
		return buf
	}

	var flags byte
	if ins.OutOfNetwork {
		flags |= (1 << 7)
	}
	// assume program_splice is set;
	// we do not support the deprecated component_count mode.
	flags |= (1 << 6)
	if ins.Duration != nil {
		flags |= (1 << 5)
	}
	if ins.Immediate {
		flags |= (1 << 4)
	}
	if ins.idCompliance {
		flags |= (1 << 3)
	}
	// toggle remaining 3 reserved bits.
	flags |= 0x07
	buf = append(buf, flags)

	if ins.SpliceTime != nil && !ins.Immediate {
		b := encodeSpliceTime(*ins.SpliceTime)
		buf = append(buf, b[:]...)
	}

	if ins.Duration != nil {
		b := packBreakDuration(ins.Duration)
		buf = append(buf, b[:]...)
	}
	buf = append(buf, byte(ins.ProgramID>>8))
	buf = append(buf, byte(ins.ProgramID))
	buf = append(buf, byte(ins.AvailNum), byte(ins.AvailExpected))
	return buf
}

func encodeSpliceTime(ticks uint64) [5]byte {
	pts := toPTS(ticks)
	// set time_specified_flag
	pts[0] |= (1 << 7)
	// toggle 6 reserved bits, so that we match the spec.
	pts[0] |= 0x7e
	return pts
}
