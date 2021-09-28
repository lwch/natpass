package pool

import (
	"crypto/tls"
	"math/rand"
	"natpass/code/client/global"
	"natpass/code/network"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// Pool connection pool
type Pool struct {
	sync.RWMutex
	cfg   *global.Configure
	conns map[uint32]*Conn
	count int
	idx   uint32
}

// New create connection pool
func New(cfg *global.Configure) *Pool {
	return &Pool{
		cfg:   cfg,
		conns: make(map[uint32]*Conn, cfg.Links),
		count: cfg.Links,
		idx:   rand.Uint32(), // random initialize for fast restart
	}
}

func (p *Pool) getConns() []*Conn {
	ret := make([]*Conn, 0, len(p.conns))
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
	if len(id) > 0 {
		for _, conn := range conns {
			if conn.hasLink(id[0]) {
				return conn
			}
		}
	}
	if len(conns) >= p.count {
		p.Lock()
		conn := conns[int(p.idx)%len(conns)]
		p.idx++
		p.Unlock()
		return conn
	}

	idx := atomic.AddUint32(&p.idx, 1)
	conn := p.connect(idx)
	if conn == nil {
		return nil
	}
	c := newConn(p, conn, idx)

	p.Lock()
	p.conns[c.Idx] = c
	p.Unlock()
	return c
}

func (p *Pool) connect(idx uint32) *network.Conn {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("connect error: %v", err)
		}
	}()
	conn, err := tls.Dial("tcp", p.cfg.Server, nil)
	runtime.Assert(err)
	c := network.NewConn(conn)
	err = p.writeHandshake(c, p.cfg, idx)
	runtime.Assert(err)
	logging.Info("%s connected", p.cfg.Server)
	return c
}

func (p *Pool) writeHandshake(conn *network.Conn, cfg *global.Configure, idx uint32) error {
	var msg network.Msg
	msg.XType = network.Msg_handshake
	msg.From = p.cfg.ID
	msg.FromIdx = idx
	msg.To = "server"
	msg.Payload = &network.Msg_Hsp{
		Hsp: &network.HandshakePayload{
			Enc: cfg.Enc[:],
		},
	}
	return conn.WriteMessage(&msg, 5*time.Second)
}

func (p *Pool) onClose(idx uint32) {
	p.Lock()
	delete(p.conns, idx)
	p.Unlock()
}

func (p *Pool) Size() int {
	return len(p.conns)
}
