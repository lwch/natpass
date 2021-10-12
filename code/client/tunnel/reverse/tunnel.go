package reverse

import (
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/tunnel"
	"net"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// Tunnel tunnel object
type Tunnel struct {
	sync.RWMutex
	Name  string
	cfg   global.Tunnel
	links map[string]*Link
}

// New new tunnel
func New(cfg global.Tunnel) *Tunnel {
	return &Tunnel{
		Name: cfg.Name,
		cfg:  cfg,
	}
}

// GetName get reverse tunnel name
func (tunnel *Tunnel) GetName() string {
	return tunnel.Name
}

// GetTypeName get reverse tunnel type name
func (tunnel *Tunnel) GetTypeName() string {
	return "reverse"
}

// GetTarget get target of this tunnel
func (tunnel *Tunnel) GetTarget() string {
	return tunnel.cfg.Target
}

// GetLinks get tunnel links
func (t *Tunnel) GetLinks() []tunnel.Link {
	ret := make([]tunnel.Link, 0, len(t.links))
	t.RLock()
	for _, link := range t.links {
		ret = append(ret, link)
	}
	t.RUnlock()
	return ret
}

// Handle handle tunnel
func (tunnel *Tunnel) Handle(pool *pool.Pool) {
	if tunnel.cfg.Type == "tcp" {
		tunnel.handleTcp(pool)
	} else {
		// TODO
		func() {}()
	}
}

func (tunnel *Tunnel) handleTcp(pool *pool.Pool) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close tcp tunnel: %s, err=%v", tunnel.cfg.Name, err)
		}
	}()
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(tunnel.cfg.LocalAddr),
		Port: int(tunnel.cfg.LocalPort),
	})
	runtime.Assert(err)
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			logging.Error("accept from %s tunnel, err=%v", tunnel.cfg.Name, err)
			continue
		}

		id, err := runtime.UUID(16, "0123456789abcdef")
		if err != nil {
			conn.Close()
			logging.Error("generate link id failed, err=%v", err)
			continue
		}

		remote := pool.Get(id)
		if remote == nil {
			conn.Close()
			logging.Error("no connection available")
			continue
		}
		link := NewLink(tunnel, id, tunnel.cfg.Target, conn, remote)
		tunnel.Lock()
		tunnel.links[id] = link
		tunnel.Unlock()
		remote.SendConnectReq(id, tunnel.cfg)
		link.Forward()
	}
}

func (tunnel *Tunnel) remove(id string) {
	tunnel.Lock()
	delete(tunnel.links, id)
	tunnel.Unlock()
}
