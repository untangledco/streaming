// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or   implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package scte35 contains the structs for encoding and/or decoding SCTE-35
// signal.
package scte35

import (
	"errors"
	"math"
	"time"
)

const (
	// Reserved bits shall be set to 1
	Reserved = 0xFF
	// TicksPerSecond is the number of 90KHz ticks per second
	TicksPerSecond = 90000
	// unixEpochToGPSEpoch is the number of seconds between 1970-01-01T00:00:00Z
	// (Unix Epoch) and 1980-01-06T00:00:00Z (GPS Epoch).
	unixEpochToGPSEpoch = uint32(315964800)
)

var (
	// ErrBufferUnderflow is returned when decoding fails to fully consume the
	// provided byte array.
	ErrBufferUnderflow = errors.New("buffer underflow")
	// ErrBufferOverflow is returned when decoding requires more bytes than are
	// available.
	ErrBufferOverflow = errors.New("buffer overflow")
	// ErrUnsupportedEncoding is returned if the signal is not a base-64 encoded
	// string.
	ErrUnsupportedEncoding = errors.New("invalid or unsupported encoding")
)

/*
// Logger for emitting debug messages.
var Logger = log.New(io.Discard, "SCTE35 ", log.Ldate|log.Ltime|log.Llongfile)

// DecodeBase64 is a convenience function for decoding a base-64 string into
// a SpliceInfoSection. If an error occurs, the returned SpliceInfoSection
// will contain the results of decoding up until the error condition
// was encountered.
func DecodeBase64(s string) (*SpliceInfoSection, error) {
	sis := &SpliceInfoSection{}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return sis, ErrUnsupportedEncoding
	}
	err = sis.Decode(b)
	return sis, err
}

// DecodeHex is a convenience function for decoding a hexadecimal string into
// a SpliceInfoSection. If an error occurs, the returned SpliceInfoSection
// will contains the results of decoding up until the error condition
// was encountered.
func DecodeHex(s string) (*SpliceInfoSection, error) {
	sis := &SpliceInfoSection{}
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		return sis, ErrUnsupportedEncoding
	}
	err = sis.Decode(b)
	return sis, err
}
*/

// DurationToTicks converts a duration to 90MhZ ticks.
func DurationToTicks(d time.Duration) uint64 {
	return uint64(math.Ceil(float64(d) * TicksPerSecond / float64(time.Second)))
}

// TicksToDuration converts 90MhZ ticks to a duration.
func TicksToDuration(ticks uint64) time.Duration {
	s := float64(ticks) / float64(TicksPerSecond)
	return time.Duration(int64(s * float64(time.Second)))
}
