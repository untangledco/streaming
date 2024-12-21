// Package wav provides an encoder and decoder of WAVE data.
// https://en.wikipedia.org/wiki/WAV
package wav

import (
	"encoding/binary"
	"fmt"
	"io"
)

type File struct {
	Header    Header
	Bitstream io.Reader
}

type Header struct {
	FileSize uint32

	BlocSize       uint32
	AudioFormat    uint16
	ChannelCount   uint16
	Frequency      uint32
	BytesPerSecond uint32
	BytesPerBloc   uint16
	BitsPerSample  uint16

	DataSize uint32
}

type header struct {
	FileBlocID   [4]byte
	FileSize     uint32
	FileFormatID [4]byte

	FormatBlocID   [4]byte
	BlocSize       uint32
	AudioFormat    uint16
	ChannelCount   uint16
	Frequency      uint32
	BytesPerSecond uint32
	BytesPerBloc   uint16
	BitsPerSample  uint16

	DataBlocID [4]byte
	DataSize   uint32
}

var (
	fileBlocID   = [4]byte{'R', 'I', 'F', 'F'}
	fileFormatID = [4]byte{'W', 'A', 'V', 'E'}
)

var formatBlocID = [4]byte{'f', 'm', 't', ' '}

var dataBlocID = [4]byte{'d', 'a', 't', 'a'}

func readHeader(rd io.Reader) (*header, error) {
	var h header
	if err := binary.Read(rd, binary.LittleEndian, &h); err != nil {
		return nil, err
	}
	if h.FileBlocID != fileBlocID {
		return nil, fmt.Errorf("bad file block id %x", h.FileBlocID)
	} else if h.FileFormatID != fileFormatID {
		return nil, fmt.Errorf("bad file format id %x", h.FileFormatID)
	} else if h.FormatBlocID != formatBlocID {
		return nil, fmt.Errorf("bad format block id %x", h.FileFormatID)
	} else if h.DataBlocID != dataBlocID {
		return nil, fmt.Errorf("bad data block id %x", h.DataBlocID)
	}
	return &h, nil
}

func EncodeHeader(hdr Header) [44]byte {
	var buf [44]byte
	h := header{
		fileBlocID,
		hdr.FileSize,
		fileFormatID,
		formatBlocID,
		hdr.BlocSize,
		hdr.AudioFormat,
		hdr.ChannelCount,
		hdr.Frequency,
		hdr.BytesPerSecond,
		hdr.BytesPerBloc,
		hdr.BitsPerSample,
		dataBlocID,
		hdr.DataSize,
	}
	binary.Encode(buf[:], binary.LittleEndian, h)
	return buf
}

func ReadFile(r io.Reader) (*File, error) {
	h, err := readHeader(r)
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	return &File{
		Header: Header{
			h.FileSize,
			h.BlocSize,
			h.AudioFormat,
			h.ChannelCount,
			h.Frequency,
			h.BytesPerSecond,
			h.BytesPerBloc,
			h.BitsPerSample,
			h.DataSize,
		},
		Bitstream: r,
	}, nil
}
