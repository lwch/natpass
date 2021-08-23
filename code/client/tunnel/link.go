package tunnel

import (
	"bytes"
	"io"
	"natpass/code/network"
	"net"

	"github.com/lwch/logging"
)

type Link struct {
	tunnel *Tunnel
	ID     string // link id
	target string // remote client id
	conn   net.Conn
	write  chan *network.Msg
	OnWork chan struct{}
	closed bool
}

func newLink(id, target string, tunnel *Tunnel, conn net.Conn, write chan *network.Msg) *Link {
	logging.Info("create link %s: %s", tunnel.Name, id)
	return &Link{
		tunnel: tunnel,
		ID:     id,
		target: target,
		conn:   conn,
		write:  write,
		OnWork: make(chan struct{}),
		closed: false,
	}
}

// Close close link
func (link *Link) Close() {
	if link.closed {
		return
	}
	logging.Info("disconnect tunnel %s on link %s", link.tunnel.Name, link.ID)
	link.closed = true
	err := link.conn.Close()
	if err == nil {
		link.sendDisconnect(link.ID, link.target)
	}
	link.tunnel.Close(link)
}

func (link *Link) loop() {
	defer link.Close()
	<-link.OnWork
	buf := make([]byte, 32*1024)
	for {
		n, err := link.conn.Read(buf)
		if err != nil {
			logging.Error("read data on tunnel %s link %s failed, err=%v", link.tunnel.Name, link.ID, err)
			return
		}
		if n == 0 {
			continue
		}
		logging.Debug("link %s on tunnel %s read from local %d bytes", link.ID, link.tunnel.Name, n)
		link.sendData(link.ID, link.target, buf[:n])
	}
}

// WriteData write data from remote
func (link *Link) WriteData(data []byte) error {
	logging.Debug("link %s on tunnel %s write from remote %d bytes", link.ID, link.tunnel.Name, len(data))
	_, err := io.Copy(link.conn, bytes.NewReader(data))
	return err
}
