package pool

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"natpass/code/client/global"
	"natpass/code/network"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// Pool connection pool
type Pool struct {
	sync.RWMutex
	cfg   *global.Configure
	conns map[string]*Conn
	count int
	idx   int
}

// New create connection pool
func New(cfg *global.Configure) *Pool {
	return &Pool{
		cfg:   cfg,
		conns: make(map[string]*Conn, cfg.Links),
		count: cfg.Links,
		idx:   0,
	}
}

func (p *Pool) getConns() []*Conn {
	ret := make([]*Conn, len(p.conns))
	p.RLock()
	for _, conn := range p.conns {
		ret = append(ret, conn)
	}
	p.RUnlock()
	return ret
}

// Get get connection
func (p *Pool) Get(id ...string) *Conn {
	conns := p.getConns()
	if len(conns) >= p.count {
		p.Lock()
		conn := conns[p.idx%len(conns)]
		p.idx++
		p.Unlock()
		if len(id) > 0 {
			conn.addLink(id[0])
		}
		return conn
	}

	conn := p.connect()
	if conn == nil {
		return nil
	}
	c := newConn(p, conn)
	if len(id) > 0 {
		c.addLink(id[0])
	}

	p.Lock()
	p.conns[c.ID] = c
	p.idx++
	p.Unlock()
	return c
}

func (p *Pool) connect() *network.Conn {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("connect error: %v", err)
		}
	}()
	conn, err := tls.Dial("tcp", p.cfg.Server, nil)
	runtime.Assert(err)
	c := network.NewConn(conn)
	err = p.writeHandshake(c, p.cfg, rand.Int())
	runtime.Assert(err)
	logging.Info("%s connected", p.cfg.Server)
	return c
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

func (p *Pool) onClose(id string) {
	p.Lock()
	delete(p.conns, id)
	p.Unlock()
}
