package pool

import (
	"context"
	"crypto/tls"
	"natpass/code/client/global"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// Pool connection pool
type Pool struct {
	sync.RWMutex
	count        int
	writeChannel chan *network.Msg
	tunnels      map[string]*tunnel.Tunnel // tunnel name => tunnel
	links        map[string]*tunnel.Link   // link id => link
}

// New create connection pool
func New(count int) *Pool {
	return &Pool{
		count:        count,
		writeChannel: make(chan *network.Msg, 100),
		tunnels:      make(map[string]*tunnel.Tunnel),
		links:        make(map[string]*tunnel.Link),
	}
}

// WriteChan get write channel
func (p *Pool) WriteChan() chan *network.Msg {
	return p.writeChannel
}

// Loop main loop
func (p *Pool) Loop(cfg *global.Configure) {
	for i := 0; i < p.count; i++ {
		go func() {
			for {
				p.connect(cfg)
			}
		}()
	}
	select {}
}

// LinkClose on close link
func (p *Pool) LinkClose(name, id string) {
	p.Lock()
	defer p.Unlock()
	if tunnel, ok := p.tunnels[name]; ok {
		if len(tunnel.GetLinks()) == 0 {
			delete(p.tunnels, name)
		}
	}
	delete(p.links, id)
}

// connect connect server
func (p *Pool) connect(cfg *global.Configure) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("connect error: %v", err)
		}
	}()
	conn, err := tls.Dial("tcp", cfg.Server, nil)
	runtime.Assert(err)
	c := network.NewConn(conn)
	defer c.Close()
	err = p.writeHandshake(c, cfg)
	runtime.Assert(err)
	logging.Info("%s connected", cfg.Server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case msg := <-p.writeChannel:
				msg.From = cfg.ID
				c.WriteMessage(msg, time.Second)
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		msg, err := c.ReadMessage(time.Second)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			logging.Error("read message: %v", err)
			return
		}
		logging.Info("recv: %s", msg.GetXType().String())
		switch msg.GetXType() {
		case network.Msg_connect_req:
			p.handleConnectReq(c, msg.GetFrom(), msg.GetTo(), msg.GetCreq())
		case network.Msg_connect_rep:
			p.handleConnectRep(msg.GetCrep())
		case network.Msg_disconnect:
			p.handleDisconnect(msg.GetXDisconnect())
		case network.Msg_forward:
			p.handleData(msg.GetXData())
		}
	}
}

// Add add tunnel
func (p *Pool) Add(tunnel *tunnel.Tunnel) {
	p.Lock()
	defer p.Unlock()
	p.tunnels[tunnel.Name] = tunnel
	for _, link := range tunnel.GetLinks() {
		p.links[link.ID] = link
	}
}

// AddLink add link
func (p *Pool) AddLink(link *tunnel.Link) {
	p.links[link.ID] = link
}

func (p *Pool) writeHandshake(conn *network.Conn, cfg *global.Configure) error {
	var msg network.Msg
	msg.XType = network.Msg_handshake
	msg.From = cfg.ID
	msg.To = "server"
	msg.Payload = &network.Msg_Hsp{
		Hsp: &network.HandshakePayload{
			Enc: cfg.Enc[:],
		},
	}
	return conn.WriteMessage(&msg, 5*time.Second)
}
