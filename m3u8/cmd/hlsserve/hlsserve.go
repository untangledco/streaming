// Command hlsserve serves a live HLS stream from MPEG-TS video received over TCP.
// The options are:
//
//	-l address
//		Listen for MPEG-TS streams on address, in host:port format. The
//		default is ":9000".
//	-h address
//		Listen for HTTP clients on address, in host:port format. The
//		default is ":8080".
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/untangledco/streaming/m3u8"
	"github.com/untangledco/streaming/mpegts"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("hlsserve: ")
	var err error
	cacheDir, err = os.UserCacheDir()
	if err != nil {
		log.Fatalln("user cache dir:", err)
	}
	cacheDir = filepath.Join(cacheDir, "hlsserve")
}

// rule of thumb for UDP transport
const maxTSBytes int = 7 * mpegts.PacketSize

const segmentDuration = 4 * time.Second

var cacheDir string
var sequence int

func removeOld(dir string, maxAge time.Duration) error {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, dent := range ents {
		if !strings.HasSuffix(dent.Name(), ".ts") {
			continue
		}
		stat, err := dent.Info()
		if err != nil {
			return err
		}
		if time.Since(stat.ModTime()) > maxAge {
			if err := os.Remove(filepath.Join(dir, dent.Name())); err != nil {
				return err
			}
			sequence++
		}
	}
	return nil
}

func writeSegments(dir string, r io.Reader, ch <-chan time.Time) error {
	var segment int
	segments := &bytes.Buffer{}
	sc := mpegts.NewScanner(r)
	for {
		select {
		case <-ch:
			s := fmt.Sprintf("%04d.ts", segment)
			fname := path.Join(dir, s)
			if err := os.WriteFile(fname, segments.Bytes(), 0644); err != nil {
				return err
			}
			segments.Reset()
			segment++
		default:
			if sc.Scan() {
				if err := mpegts.Encode(segments, sc.Packet()); err != nil {
					return fmt.Errorf("segment %d: encode packet: %w", segment, err)
				}
			}
			if sc.Err() != nil {
				return fmt.Errorf("segment %d: scan: %w", segment, sc.Err())
			}
		}
	}
	fmt.Fprintln(os.Stderr, "shouldn't really be returning here...?")
	return nil
}

func makePlaylist(dir string) (*m3u8.Playlist, error) {
	names, err := filepath.Glob(filepath.Join(dir, "*.ts"))
	if err != nil {
		return nil, fmt.Errorf("find segments: %w", err)
	}
	playlist := &m3u8.Playlist{
		Version:        7,
		TargetDuration: segmentDuration,
		Sequence:       sequence,
	}
	for _, name := range names {
		seg := m3u8.Segment{URI: path.Base(name), Duration: playlist.TargetDuration}
		playlist.Segments = append(playlist.Segments, seg)
	}
	return playlist, nil
}

const usage string = "usage: hlsserve dir"

func servePlaylist(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		playlist, err := makePlaylist(dir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", m3u8.MimeType)
		if err := m3u8.Encode(w, playlist); err != nil {
			log.Printf("encode playlist: %v", err)
		}
	}
}

func setCache(seconds int, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", seconds))
		next.ServeHTTP(w, req)
	}
}

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	} else if len(os.Args) == 2 {
		cacheDir = os.Args[1]
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		ticker := time.NewTicker(segmentDuration)
		if err := writeSegments(cacheDir, conn, ticker.C); err != nil {
			log.Fatalln("write segments:", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(segmentDuration)
		for {
			select {
			case <-ticker.C:
				if err := removeOld(cacheDir, 3*8*segmentDuration); err != nil {
					log.Println("remove old segments:", err)
				}
			}
		}
	}()

	http.Handle("/playlist.m3u8", servePlaylist(cacheDir))
	fsys := http.FileServer(http.FS(os.DirFS(cacheDir)))
	http.Handle("/", setCache(60, fsys))
	log.Fatal(http.ListenAndServe(":8000", nil))
}
