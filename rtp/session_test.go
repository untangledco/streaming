package rtp

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

// fakePlayer is a basic RTP packet receiver which discards packet
// payloads. It verifies the stream of packets by inspecting the packet header,
// thought it requires packets are received in order.
type fakePlayer struct {
	conn       net.PacketConn
	decoded    chan *Packet
	clock      int
	typ        PayloadType
	syncSource uint32
}

func (fp *fakePlayer) render(ch chan error) {
	var prev Packet
	for p := range fp.decoded {
		if prev.Payload == nil {
			fp.typ = p.Header.Type
			fp.syncSource = p.Header.SyncSource
			prev = *p
			continue
		}
		if p.Header.Version != VersionRFC3550 {
			ch <- fmt.Errorf("bad version %d, want %d", p.Header.Version, VersionRFC3550)
		}
		if p.Header.Type != fp.typ {
			ch <- fmt.Errorf("unexpected payload type %d, want %d", p.Header.Type, fp.typ)
		}
		if p.Header.Sequence != prev.Header.Sequence+1 {
			ch <- fmt.Errorf("bad packet sequence: previous %d, current %d", prev.Header.Sequence, p.Header.Sequence)
		}
		// TODO(otl): check timestamp is expected based on fp.clock.
		if p.Header.SyncSource != fp.syncSource {
			ch <- fmt.Errorf("unexpected sync source %d, want %d", p.Header.SyncSource, fp.syncSource)
		}
		// TODO(otl): check payload is expected? non-nil?
		prev = *p
	}
}

func (fp *fakePlayer) receive(ch chan error) {
	go fp.render(ch)
	buf := make([]byte, 1492)
	for {
		if err := fp.conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
			ch <- err
		}
		n, _, err := fp.conn.ReadFrom(buf)
		if errors.Is(err, net.ErrClosed) {
			break
		} else if err != nil {
			ch <- err
			continue
		}
		var p Packet
		if err := Unmarshal(buf[:n], &p); err != nil {
			ch <- fmt.Errorf("unmarshal packet: %w", err)
			continue
		}
		fp.decoded <- &p
	}
	close(fp.decoded)
}

func (fp *fakePlayer) stop() error {
	return fp.conn.Close()
}

// textPackets returns a channel that sends count Packets through ch every dur...
func textPackets(dur time.Duration, count int) chan Packet {
	ch := make(chan Packet)
	go func() {
		typ := DynamicPayloadType()
		ticker := time.NewTicker(dur)
		var i int
		for t := range ticker.C {
			ch <- Packet{
				Header{Type: typ},
				[]byte(t.Format(time.RFC3339Nano)),
			}
			i++
			if i == count {
				ticker.Stop()
				close(ch)
				break
			}
		}
	}()
	return ch
}

func TestSession(t *testing.T) {
	ln, err := net.ListenPacket("udp", "[::1]:0")
	if err != nil {
		t.Fatal(err)
	}
	player := fakePlayer{
		conn:    ln,
		decoded: make(chan *Packet),
		clock:   ClockText,
	}

	errs := make(chan error)
	go player.receive(errs)

	session, err := Dial("udp", ln.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	session.Clock = ClockText
	pchan := textPackets(40*time.Millisecond, 25)
	for {
		select {
		case err := <-errs:
			t.Error(err)
		case p, ok := <-pchan:
			if !ok {
				if err := player.stop(); err != nil {
					t.Errorf("stop fake player: %v", err)
				}
				return
			}
			if err := session.Transmit(&p); err != nil {
				t.Errorf("transmit: %v", err)
				continue
			}
		}
	}
}
