// Command avhelp serves...
package main

import (
	"embed"
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/untangledco/streaming/rtp"
)

//go:embed template
var content embed.FS

type server struct {
	templates *template.Template
	pconn     net.PacketConn
	buf       [5]rtp.Packet
}

func (srv *server) serveRTPInfo(w http.ResponseWriter, req *http.Request) {
	if err := srv.templates.ExecuteTemplate(w, "rtp.html", srv.buf); err != nil {
		log.Println("serve %s: render rtp.html: %v", req.URL, err)
	}
}

func (srv *server) receive() {
	buf := make([]byte, 1360)
	var i int
	for {
		n, _, err := srv.pconn.ReadFrom(buf)
		if err != nil {
			log.Printf("receive packet: %v", err)
		}
		var packet rtp.Packet
		if err := rtp.Unmarshal(buf[:n], &packet); err != nil {
			log.Printf("unmarshal packet: %v", err)
		}
		if i == 5 {
			srv.buf[0] = packet
			i = 1
		} else {
			srv.buf[i] = packet
			i++
		}
	}
}

func main() {
	tmpl, err := template.ParseFS(content, "template/*.html")
	if err != nil {
		log.Fatal(err)
	}
	pc, err := net.ListenPacket("udp", ":6969")
	if err != nil {
		log.Fatal(err)
	}

	srv := &server{
		templates: tmpl,
		pconn:     pc,
	}
	go srv.receive()

	http.HandleFunc("/rtp/", srv.serveRTPInfo)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
