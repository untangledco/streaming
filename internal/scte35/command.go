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
	// Number of ticks of a 90KHz clock.
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
		return nil, nil
	case SpliceSchedule:
		b, err := packEvents(c.Schedule)
		if err != nil {
			return b, fmt.Errorf("pack events: %w", err)
		}
		return b, nil
	case SpliceInsert:
		b := encodeInsert(c.Insert)
		return b, nil
	case TimeSignal:
		if c.TimeSignal == nil {
			return nil, fmt.Errorf("command type is %s, but nil TimeSignal value set", c.Type)
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
	// Indicates the event's ID is prepared in the method
	// described in SCTE 35 section 9.3.3.
	IDCompliance bool

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
	if e.IDCompliance {
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
	i := 4
	for j := range c.Data {
		buf[i] = c.Data[j]
		i++
	}
	return buf
}

// Insert represents the splice_insert command
// as specified in SCTE 35 section 9.7.3.
type Insert struct {
	ID                uint32
	Cancel            bool
	OutOfNetwork      bool
	Immediate         bool
	EventIDCompliance bool
	// Number of ticks of a 90KHz clock.
	SpliceTime    *uint64
	Duration      *BreakDuration
	ProgramID     uint16
	AvailNum      uint8
	AvailExpected uint8
}

func encodeInsert(ins *Insert) []byte {
	buf := make([]byte, 4+1) // uint32 + 1 byte
	binary.BigEndian.PutUint32(buf[:4], ins.ID)
	if ins.Cancel {
		buf[4] |= (1 << 7)
	}
	// toggle remaining reserved 7 bits
	buf[4] |= 0x7f

	if !ins.Cancel {
		buf = append(buf, 0x00)
		if ins.OutOfNetwork {
			buf[5] |= (1 << 7)
		}
		// assume program_splice is set;
		// we do not support the deprecated component_count mode.
		buf[5] |= (1 << 6)
		if ins.Duration != nil {
			buf[5] |= (1 << 5)
		}
		if ins.Immediate {
			buf[5] |= (1 << 4)
		}
		if ins.EventIDCompliance {
			buf[5] |= (1 << 3)
		}
		// toggle remaining 3 reserved bits.
		buf[5] |= 0x07

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
		buf = append(buf, byte(ins.AvailNum))
		buf = append(buf, byte(ins.AvailExpected))
	}
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
