package tunnel

import (
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"net"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type Tunnel struct {
	Name string
	cfg  global.Tunnel
}

// New new tunnel
func New(cfg global.Tunnel) *Tunnel {
	return &Tunnel{
		Name: cfg.Name,
		cfg:  cfg,
	}
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
		remote.SendConnectReq(id, tunnel.cfg)
		link.Forward()
	}
}
