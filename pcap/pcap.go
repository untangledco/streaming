// Package pcap implements decoding and encoding of libcap's savefile
// format described in [pcap-savefile(5)].
//
// [pcap-savefile(5)]: https://www.tcpdump.org/manpages/pcap-savefile.5.txt
package pcap

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

const (
	magicLittleEndian uint32 = 0xa1b2c3d4
	magicBigEndian    uint32 = 0xd4c3b2a1
)

var version = [2]uint16{2, 4}

type GlobalHeader struct {
	_       int32 // leftover from old pcap versions
	_       int32 // leftover from old pcap versions
	SnapLen uint32
	Network uint32
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

// Decode returns a decoded savefile packet capture from rd.
func Decode(rd io.Reader) (*File, error) {
	var magic uint32
	if err := binary.Read(rd, binary.NativeEndian, &magic); err != nil {
		return nil, fmt.Errorf("read magic number: %w", err)
	}
	if magic != magicLittleEndian && magic != magicBigEndian {
		return nil, fmt.Errorf("unknown magic number %#x", magic)
	}

	var v [2]uint16
	if err := binary.Read(rd, binary.LittleEndian, &v); err != nil {
		return nil, fmt.Errorf("read pcap version: %w", err)
	}
	if v != version {
		return nil, fmt.Errorf("unsupported version %d.%d", v[0], v[1])
	}

	var gheader GlobalHeader
	if err := binary.Read(rd, binary.LittleEndian, &gheader); err != nil {
		return nil, fmt.Errorf("read pcap version: %w", err)
	}

	var packets []Packet
	for i := 1; ; i++ {
		var h header
		err := binary.Read(rd, binary.LittleEndian, &h)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("packet %d: read header: %w", i, err)
		}
		var p Packet
		p.Header = Header{
			Time:    time.Unix(int64(h.Seconds), int64(h.SubSeconds)*1000),
			OrigLen: h.OrigLen,
		}

		p.Data = make([]byte, h.InclLen)
		if _, err = io.ReadFull(rd, p.Data); err != nil {
			return nil, fmt.Errorf("packet %d: read data: %w", i, err)
		}

		packets = append(packets, p)
	}

	return &File{gheader, packets}, nil
}

// Encode writes a savefile-encoded representation of file to w.
func Encode(w io.Writer, file *File) (n int64, err error) {
	b := make([]byte, 4+4) // magic + version
	binary.NativeEndian.PutUint32(b, magicLittleEndian)
	binary.NativeEndian.PutUint16(b[4:6], version[0])
	binary.NativeEndian.PutUint16(b[6:8], version[1])
	nn, err := w.Write(b)
	n += int64(nn)
	if err != nil {
		return n, fmt.Errorf("magic header: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, &file.Header); err != nil {
		return n, fmt.Errorf("global header: %w", err)
	}
	n += (4*4) // 4 32-bit ints in GlobalHeader

	for i, p := range file.Packets {
		sec, nsec := timestamp(p.Header.Time)
		h := header{
			Seconds:    sec,
			SubSeconds: nsec / 1000,
			InclLen:    uint32(len(p.Data)),
			OrigLen:    p.Header.OrigLen,
		}
		if err := binary.Write(w, binary.LittleEndian, h); err != nil {
			return n, fmt.Errorf("packet %d: header: %v", i, err)
		}
		n += 4*4 // 4 uint32s in header
		nn, err := w.Write(p.Data)
		n += int64(nn)
		if err != nil {
			return n, fmt.Errorf("packet %d: data: %v", i, err)
		}
	}
	return n, nil
}

func timestamp(t time.Time) (seconds, nanoSeconds uint32) {
	seconds = uint32(t.Unix())
	nanoSeconds = uint32(t.UnixNano() - t.Unix()*1e9)
	return
}
