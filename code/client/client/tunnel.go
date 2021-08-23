package client

import (
	"bytes"
	"context"
	"io"
	"net"

	"github.com/lwch/logging"
)

type link struct {
	cli    *Client
	id     string // link id
	name   string // tunnel name
	target string // remote client id
	c      net.Conn
}

func newLink(id, name, target string, cli *Client, conn net.Conn) *link {
	logging.Info("create link %s: %s", name, id)
	return &link{
		cli:    cli,
		id:     id,
		name:   name,
		target: target,
		c:      conn,
	}
}

func (l *link) close() {
	logging.Info("disconnect tunnel %s on link %s", l.name, l.id)
	err := l.c.Close()
	if err == nil {
		l.cli.disconnect(l.id, l.target)
	}
}

func (l *link) loop(ctx context.Context) {
	defer l.cli.closeLink(l)
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		n, err := l.c.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}
		logging.Debug("link %s on tunnel %s read from local %d bytes", l.id, l.name, n)
		l.cli.send(l.id, l.target, buf[:n])
	}
}

func (l *link) write(data []byte) error {
	logging.Debug("link %s on tunnel %s write from remote %d bytes", l.id, l.name, len(data))
	_, err := io.Copy(l.c, bytes.NewReader(data))
	return err
}
