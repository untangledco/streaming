package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/untangledco/streaming/m3u8"
)

const usage string = "usage: hlsproxy url"

func init() {
	log.SetFlags(0)
	log.SetPrefix("hlssrv: ")
}

func servePlaylist(p *m3u8.Playlist) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "nope", http.StatusMethodNotAllowed)
		}
		log.Println(req.Method, req.URL)
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		if err := m3u8.Encode(w, p); err != nil {
			log.Println("encode playlist:", err)
		}
	}
}

func injectBreak(plist *m3u8.Playlist, dur time.Duration, schedule int) error {
	// ticks := (dur * time.Second) / 90000
	var money []m3u8.Segment
	for i, seg := range plist.Segments {
		if i%schedule != 0 || i == 0 {
			money = append(money, seg)
			continue
		}
		adbreak := m3u8.Segment{
			URI:           "https://test-streams.mux.dev/test_001/stream_1000k_48k_640x360_000.ts",
			Duration:      dur,
			Discontinuity: true,
		}
		seg.Discontinuity = true
		money = append(money, adbreak, seg)
	}
	plist.Segments = money
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(127)
	}
	link, err := url.Parse(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Get(link.String())
	if err != nil {
		log.Fatal("get playlist:", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal("get playlist: remote status:", resp.Status)
	}
	source, err := m3u8.ParsePlaylist(resp.Body)
	if err != nil {
		log.Fatal("parse playlist:", err)
	}
	link.Path = path.Dir(link.Path) + "/"
	for i := range source.Media {
		source.Media[i].URI = link.String() + source.Media[i].URI
	}
	for i := range source.Variants {
		source.Variants[i].URI = link.String() + source.Variants[i].URI
	}
	for i := range source.Segments {
		source.Segments[i].URI = link.String() + source.Segments[i].URI
	}
	if err := injectBreak(source, 10*time.Second, 8); err != nil {
		log.Fatalf("inject ad break: %v", err)
	}
	http.HandleFunc("/", servePlaylist(source))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
