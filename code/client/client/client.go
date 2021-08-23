package client

import (
	"context"
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

// Client client
type Client struct {
	sync.RWMutex
	cfg   *global.Configure
	conn  *network.Conn
	links map[string]*link
}

// New create client
func New(cfg *global.Configure, conn *network.Conn) *Client {
	return &Client{
		cfg:   cfg,
		conn:  conn,
		links: make(map[string]*link),
	}
}

// Run main loop
func (c *Client) Run() {
	err := c.writeHandshake()
	runtime.Assert(err)
	logging.Info("%s connected", c.cfg.Server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go c.keepalive(ctx)

	for _, t := range c.cfg.Tunnels {
		if t.Type == "tcp" {
			go c.handleTcpTunnel(ctx, t)
		} else {
			go c.handleUdpTunnel(ctx, t)
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
			c.handleConnect(ctx, msg.GetFrom(), msg.GetTo(), msg.GetCreq())
		case network.Msg_disconnect:
			c.handleDisconnect(msg.GetXDisconnect())
		case network.Msg_forward:
			c.handleData(msg.GetXData())
		}
	}
}

// handleTcpTunnel local listen to tcp tunnel
func (c *Client) handleTcpTunnel(ctx context.Context, t global.Tunnel) {
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
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, err := l.Accept()
		if err != nil {
			logging.Error("accept from %s tunnel, err=%v", t.Name, err)
			continue
		}

		id, err := runtime.UUID(16, "0123456789abcdef")
		if err != nil {
			conn.Close()
			logging.Error("generate link id failed, err=%v", err)
			continue
		}

		link := newLink(id, t.Name, t.Target, c, conn)
		c.sendConnect(link.id, t)

		c.Lock()
		c.links[link.id] = link
		c.Unlock()

		go link.loop(ctx)
	}
}

func (c *Client) keepalive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		time.Sleep(10 * time.Second)
		c.sendKeepalive()
	}
}

// handleUdpTunnel local listen to udp tunnel
func (c *Client) handleUdpTunnel(ctx context.Context, t global.Tunnel) {
	// TODO
}

// handleConnect handle connect request message from remote, local dial to remomte addr
func (c *Client) handleConnect(ctx context.Context, from, to string, req *network.ConnectRequest) {
	dial := "tcp"
	if req.GetXType() == network.ConnectRequest_udp {
		dial = "udp"
	}
	conn, err := net.Dial(dial, fmt.Sprintf("%s:%d", req.GetAddr(), req.GetPort()))
	if err != nil {
		c.connectError(to, req.GetId(), err.Error())
		return
	}
	link := newLink(req.GetId(), req.GetName(), from, c, conn)
	c.Lock()
	c.links[link.id] = link
	c.Unlock()
	c.connectOK(to, req.GetId())
	go link.loop(ctx)
}

// handleDisconnect handle disconnect message from remote, this means remote connection is closed
func (c *Client) handleDisconnect(data *network.Disconnect) {
	id := data.GetId()

	c.RLock()
	tn := c.links[id]
	c.RUnlock()

	if tn != nil {
		tn.close()

		c.Lock()
		delete(c.links, id)
		c.Unlock()
	}
}

// handleData handle forward data message, write data to local connection
func (c *Client) handleData(data *network.Data) {
	id := data.GetLid()
	c.RLock()
	tn := c.links[id]
	c.RUnlock()
	if tn == nil {
		logging.Error("link %s not found", id)
		return
	}
	tn.write(data.GetData())
}

func (c *Client) closeLink(l *link) {
	l.close()
	c.Lock()
	delete(c.links, l.id)
	c.Unlock()
}
