package client

import (
	"natpass/code/client/global"
	"natpass/code/network"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type Client struct {
	sync.RWMutex
	cfg     *global.Configure
	conn    *network.Conn
	tunnels map[string]*tunnel
}

func New(cfg *global.Configure, conn *network.Conn) *Client {
	return &Client{
		cfg:     cfg,
		conn:    conn,
		tunnels: make(map[string]*tunnel),
	}
}

func (c *Client) Run() {
	err := c.writeHandshake()
	runtime.Assert(err)
	logging.Info("%s connected", c.cfg.Server)

	for _, t := range c.cfg.Tunnels {
		if t.Type == "tcp" {
			go c.handleTcpTunnel(t)
		} else {
			go c.handleUdpTunnel(t)
		}
	}

	for {
		msg, err := c.conn.ReadMessage(time.Second)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			logging.Error("read message: %v", err)
			return
		}
		logging.Info("read message: %s", msg.String())
	}
}

func (c *Client) writeHandshake() error {
	var msg network.Msg
	msg.XType = network.Msg_handshake
	msg.From = c.cfg.ID
	msg.To = "server"
	msg.Payload = &network.Msg_Hsp{
		Hsp: &network.HandshakePayload{
			Enc: c.cfg.Enc[:],
		},
	}
	return c.conn.WriteMessage(&msg, 5*time.Second)
}

func (c *Client) handleTcpTunnel(t global.Tunnel) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close tcp tunnel: %s, err=%v", t.Name, err)
		}
	}()
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(t.LocalAddr),
		Port: int(t.LocalPort),
	})
	runtime.Assert(err)
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			logging.Error("accept from %s tunnel, err=%v", t.Name, err)
			continue
		}
		tn := newTunnel(c, conn)
		c.sendConnect(tn.id, t)

		c.Lock()
		c.tunnels[tn.id] = tn
		c.Unlock()
	}
}

func (c *Client) handleUdpTunnel(t global.Tunnel) {
	// TODO
}

func (c *Client) sendConnect(id string, t global.Tunnel) {
	tp := network.ConnectRequest_tcp
	if t.Type != "tcp" {
		tp = network.ConnectRequest_udp
	}
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = t.Target
	msg.XType = network.Msg_connect_req
	msg.Payload = &network.Msg_Creq{
		Creq: &network.ConnectRequest{
			Id:    id,
			Name:  t.Name,
			XType: tp,
			Addr:  t.RemoteAddr,
			Port:  uint32(t.RemotePort),
		},
	}
	c.conn.WriteMessage(&msg, 5*time.Second)
}
