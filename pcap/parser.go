package pcap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

const (
	magicLittleEndian uint32 = 0xa1b2c3d4
	magicBigEndian    uint32 = 0xd4c3b2a1
)

type GlobalHeader struct {
	VersionMajor uint16
	VersionMinor uint16
	ThisZone     int32
	SigFigs      uint32
	SnapLen      uint32
	Network      uint32
}

type Header struct {
	Time    time.Time
	OrigLen uint32
}

type header struct {
	Seconds    uint32
	SubSeconds uint32 // micro or nanoseconds
	InclLen    uint32
	OrigLen    uint32
}

type Packet struct {
	Header Header
	Data   []byte
}

type File struct {
	Header  GlobalHeader
	Packets []Packet
}

func decode(reader io.Reader) (*File, error) {
	var magic uint32
	if err := binary.Read(reader, binary.NativeEndian, &magic); err != nil {
		return nil, fmt.Errorf("read magic number: %w", err)
	}
	if magic != magicLittleEndian && magic != magicBigEndian {
		return nil, fmt.Errorf("unknown magic number %#x", magic)
	}

	var gheader GlobalHeader
	if err := binary.Read(reader, binary.LittleEndian, &gheader); err != nil {
		return nil, fmt.Errorf("read global header: %w", err)
	}

	var packets []Packet
	for i := 1; ; i++ {
		var h header
		err := binary.Read(reader, binary.LittleEndian, &h)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("packet %d: read header: %w", i, err)
		}
		hh := Header{
			Time:    time.Unix(int64(h.Seconds), int64(h.SubSeconds)*1000),
			OrigLen: h.OrigLen,
		}

		data := make([]byte, h.InclLen)
		if _, err = io.ReadFull(reader, data); err != nil {
			return nil, fmt.Errorf("packet %d: read data: %w", i, err)
		}

		packets = append(packets, Packet{
			Header: hh,
			Data:   data,
		})
	}

	return &File{
		Header:  gheader,
		Packets: packets,
	}, nil
}

func encode(file *File) ([]byte, error) {
	b := make([]byte, 4)
	binary.NativeEndian.PutUint32(b, magicLittleEndian)
	buf := bytes.NewBuffer(b)
	if err := binary.Write(buf, binary.LittleEndian, &file.Header); err != nil {
		return nil, fmt.Errorf("global header: %v", err)
	}

	for _, p := range file.Packets {
		sec, nsec := timestamp(p.Header.Time)
		h := header{
			Seconds:    sec,
			SubSeconds: nsec / 1000,
			InclLen:    uint32(len(p.Data)),
			OrigLen:    p.Header.OrigLen,
		}
		if err := binary.Write(buf, binary.LittleEndian, h); err != nil {
			return nil, fmt.Errorf("packet header: %v", err)
		}
		if err := binary.Write(buf, binary.LittleEndian, p.Data); err != nil {
			return nil, fmt.Errorf("packet data: %v", err)
		}
	}

	return buf.Bytes(), nil
}

func timestamp(t time.Time) (seconds, nanoSeconds uint32) {
	seconds = uint32(t.Unix())
	nanoSeconds = uint32(t.UnixNano() - t.Unix()*1e9)
	return
}
