package handler

import (
	"bytes"
	"natpass/code/network"
	"natpass/code/server/global"
	"net"
	"sync"
	"time"

	"github.com/lwch/logging"
)

type Handler struct {
	sync.RWMutex
	cfg     *global.Configure
	clients map[string]*client    // client id => client
	tunnels map[string][2]*client // tunnel id => endpoints
}

func New(cfg *global.Configure) *Handler {
	return &Handler{
		cfg:     cfg,
		clients: make(map[string]*client),
		tunnels: make(map[string][2]*client),
	}
}

func (h *Handler) Handle(conn net.Conn) {
	c := network.NewConn(conn)
	var id string
	defer func() {
		if len(id) > 0 {
			logging.Info("%s disconnected", id)
		}
		c.Close()
	}()
	var err error
	for i := 0; i < 10; i++ {
		id, err = h.readHandshake(c)
		if err != nil {
			if err == errInvalidHandshake {
				logging.Error("invalid handshake from %s", c.RemoteAddr().String())
				return
			}
			logging.Error("read handshake from %s %d times, err=%v", c.RemoteAddr().String(), i+1, err)
			continue
		}
		break
	}
	if err != nil {
		return
	}
	logging.Info("%s connected", id)

	cli := newClient(h, id, c)
	h.Lock()
	h.clients[cli.id] = cli
	h.Unlock()

	defer h.closeAll(cli)

	cli.run()
}

func (h *Handler) readHandshake(c *network.Conn) (string, error) {
	msg, err := c.ReadMessage(5 * time.Second)
	if err != nil {
		return "", err
	}
	if msg.GetXType() != network.Msg_handshake {
		return "", errNotHandshake
	}
	n := bytes.Compare(msg.GetHsp().GetEnc(), h.cfg.Enc[:])
	if n != 0 {
		return "", errInvalidHandshake
	}
	return msg.GetFrom(), nil
}

func (h *Handler) onMessage(msg *network.Msg) {
	to := msg.GetTo()
	h.RLock()
	cli := h.clients[to]
	h.RUnlock()
	if cli == nil {
		logging.Error("client %s not found", to)
		return
	}
	h.msgFilter(msg)
	cli.writeMessage(msg)
}

func (h *Handler) msgFilter(msg *network.Msg) {
	from := msg.GetFrom()
	to := msg.GetTo()
	switch msg.GetXType() {
	case network.Msg_connect_rep:
		if msg.GetCrep().GetOk() {
			h.RLock()
			fromCli := h.clients[from]
			toCli := h.clients[to]
			h.RUnlock()
			id := msg.GetCrep().GetId()
			var pair [2]*client
			if fromCli != nil {
				fromCli.addTunnel(id)
				pair[0] = fromCli
			}
			if toCli != nil {
				toCli.addTunnel(id)
				pair[1] = toCli
			}
			h.Lock()
			h.tunnels[id] = pair
			h.Unlock()
		}
	case network.Msg_disconnect:

	}
}

func (h *Handler) closeAll(cli *client) {
	tunnels := cli.getTunnels()
	for _, t := range tunnels {
		h.RLock()
		pair := h.tunnels[t]
		h.RUnlock()
		if pair[0] != nil {
			pair[0].close(t)
		}
		if pair[1] != nil {
			pair[1].close(t)
		}
		h.Lock()
		delete(h.tunnels, t)
		h.Unlock()
	}
	h.Lock()
	delete(h.clients, cli.id)
	h.Unlock()
}
