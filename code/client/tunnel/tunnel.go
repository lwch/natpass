package tunnel

import (
	"fmt"
	"natpass/code/network"
	"net"

	"github.com/lwch/logging"
)

type Tunnel struct {
	name            string
	localClientID   string
	localChannelID  string
	remoteClientID  string
	remoteChannelID string
	conn            net.Conn
}

func NewListen(name, localClientID, localChannelID,
	remoteClientID, remoteChannelID string, conn net.Conn) (*Tunnel, error) {
	logging.Info("new listen: name=%s, addr=%s", name, conn.LocalAddr().String())
	return &Tunnel{
		name:            name,
		localClientID:   localClientID,
		localChannelID:  localChannelID,
		remoteClientID:  remoteClientID,
		remoteChannelID: remoteChannelID,
		conn:            conn,
	}, nil
}

func NewConnect(name, localClientID, localChannelID,
	remoteClientID, remoteChannelID string, t, addr string, port uint32) (*Tunnel, error) {
	logging.Info("new connect: name=%s, remote=%s://%s:%d", name, t, addr, port)
	conn, err := net.Dial(t, fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}
	return &Tunnel{
		name:            name,
		localClientID:   localClientID,
		localChannelID:  localChannelID,
		remoteClientID:  remoteClientID,
		remoteChannelID: remoteChannelID,
		conn:            conn,
	}, nil
}

func (t *Tunnel) Close() {
	t.conn.Close()
}

// ForwardLocal read local => write remote
func (t *Tunnel) ForwardLocal(cli network.Natpass_ForwardClient) {
	defer t.Close()
	buf := make([]byte, 32*1024)
	for {
		n, err := t.conn.Read(buf)
		if err != nil {
			logging.Error("read data from local failed, name=%s, err=%v", t.name, err)
			return
		}
		err = cli.Send(&network.Data{
			From: t.localClientID,
			To:   t.remoteClientID,
			Cid:  t.remoteChannelID,
			Data: buf[:n],
		})
		if err != nil {
			logging.Error("write data to remote failed, name=%s, err=%v", t.name, err)
			return
		}
	}
}

func (t *Tunnel) WriteLocal(data []byte) (int, error) {
	return t.conn.Write(data)
}
