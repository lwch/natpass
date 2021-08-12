package client

import (
	"bytes"
	"io"
	"net"
)

type tunnel struct {
	cli    *Client
	id     string
	target string
	c      net.Conn
}

func newTunnel(id, target string, cli *Client, conn net.Conn) *tunnel {
	return &tunnel{
		cli:    cli,
		id:     id,
		target: target,
		c:      conn,
	}
}

func (tn *tunnel) close() {
	tn.c.Close()
}

func (tn *tunnel) loop() {
	defer tn.close()
	buf := make([]byte, 32*1024)
	for {
		n, err := tn.c.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}
		tn.cli.send(tn.id, tn.target, buf[:n])
	}
}

func (tn *tunnel) write(data []byte) error {
	_, err := io.Copy(tn.c, bytes.NewReader(data))
	return err
}
