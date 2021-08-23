package tunnel

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/lwch/logging"
)

type Link struct {
	tunnel *Tunnel
	ID     string // link id
	target string // remote client id
	conn   net.Conn
	OnWork chan struct{}
	closed bool
	idx    int
}

func newLink(id, target string, tunnel *Tunnel, conn net.Conn) *Link {
	logging.Info("create link %s: %s", tunnel.Name, id)
	return &Link{
		tunnel: tunnel,
		ID:     id,
		target: target,
		conn:   conn,
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
		link.tunnel.super.SendDisconnect(link.ID, link.target)
	}
	link.tunnel.Close(link)
}

func debug(data []byte, op, id string, idx int) {
	str := hex.Dump(data)
	os.MkdirAll("dump", 0755)
	ioutil.WriteFile(fmt.Sprintf("dump/%s_%s_%d.log", id, op, idx), []byte(str), 0644)
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
		debug(buf[:n], "read", link.ID, link.idx)
		link.idx++
		logging.Debug("link %s on tunnel %s read from local %d bytes", link.ID, link.tunnel.Name, n)
		link.tunnel.super.SendData(link.ID, link.target, buf[:n])
	}
}

// WriteData write data from remote
func (link *Link) WriteData(data []byte) error {
	debug(data, "write", link.ID, link.idx)
	link.idx++
	logging.Debug("link %s on tunnel %s write from remote %d bytes", link.ID, link.tunnel.Name, len(data))
	_, err := io.Copy(link.conn, bytes.NewReader(data))
	if err != nil {
		logging.Error("write data on tunnel %s link %s failed, err=%v", link.tunnel.Name, link.ID, err)
	}
	return err
}
