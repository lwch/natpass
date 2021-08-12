package client

import (
	"bytes"
	"context"
	"io"
	"net"

	"github.com/lwch/logging"
)

type tunnel struct {
	cli    *Client
	id     string
	name   string
	target string
	c      net.Conn
}

func newTunnel(id, name, target string, cli *Client, conn net.Conn) *tunnel {
	logging.Info("create tunnel %s: %s", name, id)
	return &tunnel{
		cli:    cli,
		id:     id,
		name:   name,
		target: target,
		c:      conn,
	}
}

func (tn *tunnel) close() {
	logging.Info("disconnect tunnel %s: %s", tn.name, tn.id)
	err := tn.c.Close()
	if err == nil {
		tn.cli.disconnect(tn.id, tn.target)
	}
}

func (tn *tunnel) loop(ctx context.Context) {
	defer tn.close()
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		n, err := tn.c.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}
		logging.Debug("tunnel %s read from local %d bytes", tn.name, n)
		tn.cli.send(tn.id, tn.target, buf[:n])
	}
}

func (tn *tunnel) write(data []byte) error {
	logging.Debug("tunnel %s write from remote %s bytes", tn.name, len(data))
	_, err := io.Copy(tn.c, bytes.NewReader(data))
	return err
}
