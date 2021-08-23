package pool

import (
	"context"
	"crypto/tls"
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type connectData struct {
	to   string
	id   string
	name string
	tp   network.ConnectRequestType
	addr string
	port uint16
}

type disconnectData struct {
	to string
	id string
}

type forwardData struct {
	to   string
	id   string
	data []byte
}

// Pool connection pool
type Pool struct {
	sync.RWMutex
	count           int
	writeConnect    chan connectData
	writeDisconnect chan disconnectData
	writeForward    chan forwardData
	tunnels         map[string]*tunnel.Tunnel // tunnel name => tunnel
	links           map[string]*tunnel.Link   // link id => link
}

// New create connection pool
func New(count int) *Pool {
	return &Pool{
		count:           count,
		writeConnect:    make(chan connectData),
		writeDisconnect: make(chan disconnectData),
		writeForward:    make(chan forwardData),
		tunnels:         make(map[string]*tunnel.Tunnel),
		links:           make(map[string]*tunnel.Link),
	}
}

// Loop main loop
func (p *Pool) Loop(cfg *global.Configure) {
	for i := 0; i < p.count; i++ {
		go func(i int) {
			for {
				p.connect(cfg, i)
				time.Sleep(time.Second)
			}
		}(i)
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

func (p *Pool) connect(cfg *global.Configure, idx int) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("connect error: %v", err)
		}
	}()
	conn, err := tls.Dial("tcp", cfg.Server, nil)
	runtime.Assert(err)
	c := network.NewConn(conn)
	defer c.Close()
	err = p.writeHandshake(c, cfg, idx)
	runtime.Assert(err)
	logging.Info("%s connected", cfg.Server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go p.send(ctx, c, fmt.Sprintf("%s-%d", cfg.ID, idx))
	for {
		msg, err := c.ReadMessage(time.Second)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			logging.Error("read message: %v", err)
			return
		}
		from := msg.GetFrom()
		if len(from) > 0 {
			n := strings.LastIndex(from, "-")
			if n != -1 {
				from = from[:n]
			}
			msg.From = from
		}
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

func (p *Pool) writeHandshake(conn *network.Conn, cfg *global.Configure, idx int) error {
	var msg network.Msg
	msg.XType = network.Msg_handshake
	msg.From = fmt.Sprintf("%s-%d", cfg.ID, idx)
	msg.To = "server"
	msg.Payload = &network.Msg_Hsp{
		Hsp: &network.HandshakePayload{
			Enc: cfg.Enc[:],
		},
	}
	return conn.WriteMessage(&msg, 5*time.Second)
}

func (p *Pool) send(ctx context.Context, conn *network.Conn, from string) {
	for {
		var msg network.Msg
		msg.From = from
		select {
		case m := <-p.writeConnect:
			msg.XType = network.Msg_connect_req
			msg.To = m.to
			msg.Payload = &network.Msg_Creq{
				Creq: &network.ConnectRequest{
					Id:    m.id,
					Name:  m.name,
					XType: m.tp,
					Addr:  m.addr,
					Port:  uint32(m.port),
				},
			}
		case m := <-p.writeDisconnect:
			msg.XType = network.Msg_disconnect
			msg.To = m.to
			msg.Payload = &network.Msg_XDisconnect{
				XDisconnect: &network.Disconnect{
					Id: m.id,
				},
			}
		case m := <-p.writeForward:
			msg.XType = network.Msg_forward
			msg.To = m.to
			msg.Payload = &network.Msg_XData{
				XData: &network.Data{
					Lid:  m.id,
					Data: m.data,
				},
			}
		case <-ctx.Done():
			return
		}
		conn.WriteMessage(&msg, time.Second)
	}
}
