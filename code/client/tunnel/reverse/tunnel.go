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
		Name:  cfg.Name,
		cfg:   cfg,
		links: make(map[string]*Link),
	}
}

// GetName get reverse tunnel name
func (tn *Tunnel) GetName() string {
	return tn.Name
}

// GetTypeName get reverse tunnel type name
func (tn *Tunnel) GetTypeName() string {
	return "reverse"
}

// GetTarget get target of this tunnel
func (tn *Tunnel) GetTarget() string {
	return tn.cfg.Target
}

// GetLinks get tunnel links
func (tn *Tunnel) GetLinks() []tunnel.Link {
	ret := make([]tunnel.Link, 0, len(tn.links))
	tn.RLock()
	for _, link := range tn.links {
		ret = append(ret, link)
	}
	tn.RUnlock()
	return ret
}

// GetRemote get remote target name
func (tn *Tunnel) GetRemote() string {
	return tn.cfg.Target
}

// GetPort get listen port
func (tn *Tunnel) GetPort() uint16 {
	return tn.cfg.LocalPort
}

// Handle handle tunnel
func (tn *Tunnel) Handle(pool *pool.Pool) {
	if tn.cfg.Type == "tcp" {
		tn.handleTCP(pool)
	} else {
		// TODO
		func() {}()
	}
}

func (tn *Tunnel) handleTCP(pool *pool.Pool) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close tcp tunnel: %s, err=%v", tn.cfg.Name, err)
		}
	}()
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(tn.cfg.LocalAddr),
		Port: int(tn.cfg.LocalPort),
	})
	runtime.Assert(err)
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			logging.Error("accept from %s tunnel, err=%v", tn.cfg.Name, err)
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
		link := NewLink(tn, id, tn.cfg.Target, conn, remote)
		tn.Lock()
		tn.links[id] = link
		tn.Unlock()
		remote.SendConnectReq(id, tn.cfg)
		link.Forward()
	}
}

func (tn *Tunnel) remove(id string) {
	tn.Lock()
	delete(tn.links, id)
	tn.Unlock()
}
