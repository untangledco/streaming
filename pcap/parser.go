package pcap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
	TsSec   uint32
	TsUsec  uint32
	InclLen uint32
	OrigLen uint32
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
	b := make([]byte, 4)
	if _, err := io.ReadFull(reader, b); err != nil {
		return nil, fmt.Errorf("read magic number: %w", err)
	}
	magic := binary.NativeEndian.Uint32(b)
	if magic != magicLittleEndian && magic != magicBigEndian {
		return nil, fmt.Errorf("unknown magic number %#x", magic)
	}

	var header GlobalHeader
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	var packets []Packet
	for {
		var header Header
		err := binary.Read(reader, binary.LittleEndian, &header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		data := make([]byte, header.InclLen)
		if _, err = io.ReadFull(reader, data); err != nil {
			return nil, err
		}

		packets = append(packets, Packet{
			Header: header,
			Data:   data,
		})
	}

	return &File{
		Header:  header,
		Packets: packets,
	}, nil
}

func encode(file *File) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.NativeEndian, magicLittleEndian); err != nil {
		return nil, fmt.Errorf("write magic number: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, &file.Header); err != nil {
		return nil, fmt.Errorf("global header: %v", err)
	}

	for _, packet := range file.Packets {
		if err := binary.Write(buf, binary.LittleEndian, &packet.Header); err != nil {
			return nil, fmt.Errorf("Packet Header: %v", err)
		}
		if err := binary.Write(buf, binary.LittleEndian, &packet.Data); err != nil {
			return nil, fmt.Errorf("Packet Data: %v", err)
		}
	}

	return buf.Bytes(), nil
}
