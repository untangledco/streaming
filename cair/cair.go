// Package cair provides a client to version 24.1 of the Cinegy Air
// HTTP API as documented at
// https://open.cinegy.com/products/air/24.1/cinegy-air-http-api/.
//
// Typical usage starts with creating a Client, then fetching the
// current playlist by calling Playlist on Client. For example, to print
// the names of all future items that are 3 minutes or less (e.g.
// station breaks):
//
//	client := &Client{http.DefaultClient, "http://air.example.com:5521")
//	playlist, err := client.Playlist("video")
//	if err != nil {
//		// handle error...
//	}
//	for _, it := range playlist.Items {
//		if it.ScheduledAt.Before(time.Now()) {
//			continue
//		}
//		if it.Duration <= 3*time.Minute {
//			fmt.Println(it.Name)
//		}
//	}
package cair

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"path"
)

const DefaultPort = 5521
const defaultRoot = "http://127.0.0.1:5521"

// Client is used to communicate with the Cinegy Air API.
type Client struct {
	*http.Client // http.DefaultClient if nil.
	// Root is a HTTP URL pointing to the root from the Air API is served.
	// If empty the client assumes the value "http://127.0.0.1:5521".
	Root string
}

func (c *Client) get(path string) (*http.Response, error) {
	if c.Client == nil {
		c.Client = http.DefaultClient
	}
	if c.Root == "" {
		c.Root = defaultRoot
	}
	return c.Get(c.Root + path)
}

// Playlist retrieves the current Playlist for the named device.
// Supported names include:
//   - "video"
//   - "titler_0"
//   - "logo"
//   - "cg_[0-8]" e.g. cg_0, cg_1 ...
//   - "cg_logo"
//   - "gfx_[0-8]"
//   - "audio"
//
// For all supported names, see Cinegy Air 24.1 HTTP API, chapter 1.
func (c *Client) Playlist(device string) (*Playlist, error) {
	resp, err := c.get(path.Join("/", device, "list"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("non-OK status code from engine: %s", resp.Status)
	}
	return ParsePlaylist(resp.Body)
}

func (c *Client) Status() (*Status, error) {
	resp, err := c.get("/videos/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("non-OK status code from engine: %s", resp.Status)
	}
	var st Status
	if err := xml.NewDecoder(resp.Body).Decode(&st); err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}
	return &st, nil
}

type Status struct {
	XMLName xml.Name `xml:"Status"`
	Active  struct {
		ID string `xml:"Id,attr"`
	}
	Cued struct {
		ID string `xml:"Id,attr"`
	}
	License struct {
		State string `xml:",attr"`
	}
	Output struct {
		State string `xml:",attr"`
	}
	Client clientStatus
}

type clientStatus struct {
	XMLName   xml.Name `xml:"Client"`
	Connected string   `xml:",attr"`
	Identity  string   `xml:",attr"`
}
