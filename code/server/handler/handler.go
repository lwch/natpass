package handler

import (
	"bytes"
	"natpass/code/network"
	"natpass/code/server/global"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
)

// Handler handler
type Handler struct {
	sync.RWMutex
	cfg     *global.Configure
	clients map[string]*client    // client id => client
	links   map[string][2]*client // link id => endpoints
	idx     int
}

// New create handler
func New(cfg *global.Configure) *Handler {
	return &Handler{
		cfg:     cfg,
		clients: make(map[string]*client),
		links:   make(map[string][2]*client),
		idx:     0,
	}
}

// Handle main loop
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

	// split id and index
	trimID := id
	n := strings.LastIndex(id, "-")
	if n != -1 {
		trimID = id[:n]
	}

	cli := newClient(h, id, trimID, c)
	h.Lock()
	h.clients[cli.id] = cli
	h.Unlock()

	defer h.closeAll(cli)

	cli.run()
}

func (h *Handler) connsByTrimID(id string) []*client {
	ret := make([]*client, 0, 10)
	h.RLock()
	for _, cli := range h.clients {
		if cli.trimID == id {
			ret = append(ret, cli)
		}
	}
	h.RUnlock()
	return ret
}

func (h *Handler) getClient(linkID, targetID string) *client {
	h.RLock()
	pair := h.links[linkID]
	h.RUnlock()

	if pair[0] != nil && pair[0].trimID == targetID {
		return pair[0]
	}
	if pair[1] != nil && pair[1].trimID == targetID {
		return pair[1]
	}

	conns := h.connsByTrimID(targetID)
	if len(conns) == 0 {
		return nil
	}
	conn := conns[h.idx%len(conns)]
	h.idx++
	return conn
}

// readHandshake read handshake message and compare secret encoded from md5
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

// onMessage forward message
func (h *Handler) onMessage(conn *network.Conn, msg *network.Msg) {
	if msg.GetXType() == network.Msg_keepalive {
		return
	}
	to := msg.GetTo()
	var linkID string
	switch msg.GetXType() {
	case network.Msg_connect_req:
		linkID = msg.GetCreq().GetId()
	case network.Msg_connect_rep:
		linkID = msg.GetCrep().GetId()
	case network.Msg_disconnect:
		linkID = msg.GetXDisconnect().GetId()
	case network.Msg_forward:
		linkID = msg.GetXData().GetLid()
	}
	cli := h.getClient(linkID, to)
	if cli == nil {
		logging.Error("client %s not found", to)
		return
	}
	logging.Info("link: %s, to: %s, conn: %s, cli: %s, type: %s", linkID, to, conn.RemoteAddr().String(), cli.id, msg.GetXType().String())
	h.msgHook(msg, cli)
	cli.writeMessage(msg)
}

// msgHook hook from on message
func (h *Handler) msgHook(msg *network.Msg, toCli *client) {
	from := msg.GetFrom()
	h.RLock()
	fromCli := h.clients[from]
	h.RUnlock()
	switch msg.GetXType() {
	case network.Msg_connect_req:
		id := msg.GetCrep().GetId()
		var pair [2]*client
		if fromCli != nil {
			fromCli.addLink(id)
			pair[0] = fromCli
		}
		if toCli != nil {
			toCli.addLink(id)
			pair[1] = toCli
		}
		h.Lock()
		h.links[id] = pair
		h.Unlock()
	case network.Msg_disconnect:
		if fromCli != nil {
			fromCli.removeLink(msg.GetXDisconnect().GetId())
		}
		if toCli != nil {
			toCli.removeLink(msg.GetXDisconnect().GetId())
		}
	}
}

// closeAll close all links from client
func (h *Handler) closeAll(cli *client) {
	links := cli.getLinks()
	for _, t := range links {
		h.RLock()
		pair := h.links[t]
		h.RUnlock()
		if pair[0] != nil {
			pair[0].close(t)
		}
		if pair[1] != nil {
			pair[1].close(t)
		}
		h.Lock()
		delete(h.links, t)
		h.Unlock()
	}
	h.Lock()
	delete(h.clients, cli.id)
	h.Unlock()
}
