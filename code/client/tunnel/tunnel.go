package tunnel

import (
	"natpass/code/client/global"
	"natpass/code/network"
	"net"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type Super interface {
	LinkClose(string, string) // tunnel name, link id
	WriteChan() chan *network.Msg
	AddLink(*Link)
}

// Tunnel tunnel
type Tunnel struct {
	sync.RWMutex
	super Super
	Name  string
	cfg   global.Tunnel
	links map[string]*Link
}

// New create tunnel
func New(cfg global.Tunnel, super Super) *Tunnel {
	return &Tunnel{
		super: super,
		Name:  cfg.Name,
		cfg:   cfg,
		links: make(map[string]*Link),
	}
}

func (tunnel *Tunnel) NewLink(id, target string, conn net.Conn, write chan *network.Msg) {
	link := newLink(id, target, tunnel, conn, write)
	tunnel.Lock()
	tunnel.links[link.ID] = link
	tunnel.Unlock()
}

// Handle tunnel handler
func (tunnel *Tunnel) Handle() {
	if tunnel.cfg.Type == "tcp" {
		tunnel.handleTcp()
	} else {
		// TODO
		func() {}()
	}
}

// handleTcp local listen to tcp tunnel
func (tunnel *Tunnel) handleTcp() {
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

		link := newLink(id, tunnel.cfg.Target, tunnel, conn, tunnel.super.WriteChan())
		link.sendConnect(link.ID, tunnel.cfg)

		tunnel.super.AddLink(link)
		tunnel.Lock()
		tunnel.links[link.ID] = link
		tunnel.Unlock()

		go link.loop()
	}
}

// Close close link
func (tunnel *Tunnel) Close(link *Link) {
	link.Close()
	tunnel.Lock()
	delete(tunnel.links, link.ID)
	tunnel.Unlock()
	tunnel.super.LinkClose(tunnel.Name, link.ID)
}

// GetLinks get tunnel links
func (tunnel *Tunnel) GetLinks() []*Link {
	ret := make([]*Link, 0, len(tunnel.links))
	tunnel.RLock()
	for _, l := range tunnel.links {
		ret = append(ret, l)
	}
	tunnel.RUnlock()
	return ret
}
