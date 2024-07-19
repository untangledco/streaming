package pcap

import (
	"encoding/binary"
	"errors"
	"io"
)

type GlobalHeader struct {
	MagicNumber  uint32
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
	var header GlobalHeader
	err := binary.Read(reader, binary.LittleEndian, &header)

	if err != nil {
		return nil, err
	}

	if header.MagicNumber != 0xa1b2c3d4 && header.MagicNumber != 0xd4c3b2a1 {
		return nil, errors.New("invalid magic number in pcap file")
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
		_, err = io.ReadFull(reader, data)
		if err != nil {
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
