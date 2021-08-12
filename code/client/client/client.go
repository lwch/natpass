package client

import (
	"fmt"
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
		switch msg.GetXType() {
		case network.Msg_connect_req:
			c.handleConnect(msg.GetFrom(), msg.GetTo(), msg.GetCreq())
		case network.Msg_forward:
			c.handleData(msg.GetXData())
		}
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

		id, err := runtime.UUID(16, "0123456789abcdef")
		if err != nil {
			conn.Close()
			logging.Error("generate tunnel id failed, err=%v", err)
			continue
		}

		tn := newTunnel(id, t.Target, c, conn)
		c.sendConnect(tn.id, t)

		c.Lock()
		c.tunnels[tn.id] = tn
		c.Unlock()

		go tn.loop()
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

func (c *Client) handleConnect(from, to string, req *network.ConnectRequest) {
	dial := "tcp"
	if req.GetXType() == network.ConnectRequest_udp {
		dial = "udp"
	}
	conn, err := net.Dial(dial, fmt.Sprintf("%s:%d", req.GetAddr(), req.GetPort()))
	if err != nil {
		c.connectError(to, req.GetId(), err.Error())
		return
	}
	tn := newTunnel(req.GetId(), from, c, conn)
	c.Lock()
	c.tunnels[tn.id] = tn
	c.Unlock()
	c.connectOK(to, req.GetId())
	go tn.loop()
}

func (c *Client) connectError(to, id, m string) {
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = to
	msg.XType = network.Msg_connect_rep
	msg.Payload = &network.Msg_Crep{
		Crep: &network.ConnectResponse{
			Id:  id,
			Ok:  false,
			Msg: m,
		},
	}
	c.conn.WriteMessage(&msg, time.Second)
}

func (c *Client) connectOK(to, id string) {
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = to
	msg.XType = network.Msg_connect_rep
	msg.Payload = &network.Msg_Crep{
		Crep: &network.ConnectResponse{
			Id: id,
			Ok: true,
		},
	}
	c.conn.WriteMessage(&msg, time.Second)
}

func (c *Client) send(id, target string, data []byte) {
	var msg network.Msg
	msg.From = c.cfg.ID
	msg.To = target
	msg.XType = network.Msg_forward
	msg.Payload = &network.Msg_XData{
		XData: &network.Data{
			Cid:  id,
			Data: data,
		},
	}
	c.conn.WriteMessage(&msg, time.Second)
}

func (c *Client) handleData(data *network.Data) {
	id := data.GetCid()
	c.RLock()
	tn := c.tunnels[id]
	c.RUnlock()
	if tn == nil {
		logging.Error("tunnel %s not found", id)
		return
	}
	tn.write(data.GetData())
}
