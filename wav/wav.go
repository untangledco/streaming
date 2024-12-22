// Package wav implements access to WAV (or WAVE) files.
// WAVE is a file format for storing digitised audio,
// most commonly as raw PCM signals.
//
// See also https://en.wikipedia.org/wiki/WAV
package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type File struct {
	Header    Header
	Bitstream io.Reader
}

const (
	headerLength    = 44
	extensionLength = 24
)

const (
	AudioFormatPCMInteger uint16 = 1
	AudioFormatFloat      uint16 = 3
	AudioFormatExtensible uint16 = 0xfffe
)

var (
	riffID = [4]byte{'R', 'I', 'F', 'F'}
	waveID = [4]byte{'W', 'A', 'V', 'E'}
)

var formatChunkID = [4]byte{'f', 'm', 't', ' '}

var dataChunkID = [4]byte{'d', 'a', 't', 'a'}

type Header struct {
	FileSize uint32

	AudioFormat    uint16
	ChannelCount   uint16
	Frequency      uint32
	BytesPerSecond uint32
	BytesPerBloc   uint16
	BitsPerSample  uint16
	Extension      *FormatExtension

	DataSize uint32
}

func ReadFile(r io.Reader) (*File, error) {
	h, err := readHeader(r)
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	return &File{
		Header: Header{
			h.File.Length,
			h.Format.AudioFormat,
			h.Format.ChannelCount,
			h.Format.Frequency,
			h.Format.BytesPerSecond,
			h.Format.BytesPerBloc,
			h.Format.BitsPerSample,
			h.FormatExtension,
			h.Data.Length,
		},
		Bitstream: r,
	}, nil
}

func EncodeHeader(hdr Header) []byte {
	h := header{
		File: fileChunk{riffID, hdr.FileSize, waveID},
		Format: formatChunk{
			formatChunkID,
			formatChunkLength,
			hdr.AudioFormat,
			hdr.ChannelCount,
			hdr.Frequency,
			hdr.BytesPerSecond,
			hdr.BytesPerBloc,
			hdr.BitsPerSample,
		},
		FormatExtension: hdr.Extension,
		Data:            dataChunk{dataChunkID, hdr.DataSize},
	}
	if h.Format.AudioFormat == AudioFormatExtensible {
		h.Format.Length = extendedFormatChunkLength
	}

	bcap := headerLength
	if h.Format.AudioFormat == AudioFormatExtensible {
		bcap += extensionLength
	}
	buf := bytes.NewBuffer(make([]byte, 0, bcap))
	binary.Write(buf, binary.LittleEndian, h.File)
	binary.Write(buf, binary.LittleEndian, h.Format)
	if h.FormatExtension != nil {
		binary.Write(buf, binary.LittleEndian, h.FormatExtension)
	}
	binary.Write(buf, binary.LittleEndian, h.Data)
	return buf.Bytes()
}

type header struct {
	File   fileChunk
	Format formatChunk
	// Extension holds optional extra audio format information for
	FormatExtension *FormatExtension
	Data            dataChunk
}

type fileChunk struct {
	ID       [4]byte
	Length   uint32
	FormatID [4]byte
}

const (
	formatChunkLength         = 16
	extendedFormatChunkLength = formatChunkLength + extensionLength
)

type formatChunk struct {
	ID             [4]byte
	Length         uint32
	AudioFormat    uint16
	ChannelCount   uint16
	Frequency      uint32
	BytesPerSecond uint32
	BytesPerBloc   uint16
	BitsPerSample  uint16
}

type dataChunk struct {
	ID     [4]byte
	Length uint32
}

type FormatExtension struct {
	Length      uint16
	ValidBits   uint16
	ChannelMask uint32
	SubFormat   [16]byte
}

func readHeader(rd io.Reader) (*header, error) {
	var head header
	var fchunk fileChunk
	if err := binary.Read(rd, binary.LittleEndian, &fchunk); err != nil {
		return nil, fmt.Errorf("read file chunk: %w", err)
	}
	if fchunk.ID != riffID {
		return nil, fmt.Errorf("bad RIFF id %x", fchunk.ID)
	} else if fchunk.FormatID != waveID {
		return nil, fmt.Errorf("bad WAVE file format id %x", fchunk.FormatID)
	}
	head.File = fchunk

	var fmtchunk formatChunk
	if err := binary.Read(rd, binary.LittleEndian, &fmtchunk); err != nil {
		return nil, fmt.Errorf("read file chunk: %w", err)
	}
	if fmtchunk.ID != formatChunkID {
		return nil, fmt.Errorf("bad format chunk id %x", fmtchunk.ID)
	}
	head.Format = fmtchunk

	if fmtchunk.AudioFormat == AudioFormatExtensible {
		var ext FormatExtension
		if err := binary.Read(rd, binary.LittleEndian, &ext); err != nil {
			return nil, fmt.Errorf("read format chunk extension: %w", err)
		}
		head.FormatExtension = &ext
	}

	var data dataChunk
	if err := binary.Read(rd, binary.LittleEndian, &data); err != nil {
		return nil, fmt.Errorf("read data chunk: %w", err)
	}
	head.Data = data
	return &head, nil
}
