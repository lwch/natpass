package handler

import (
	"bytes"
	"net"
	"sync"
	"time"

	"github.com/jkstack/natpass/code/network"
	"github.com/jkstack/natpass/code/server/global"
	"github.com/lwch/logging"
)

type link struct {
	id        string
	t         network.ConnectRequestType
	endPoints [2]*client
}

func (link *link) close() {
	close := func(cli *client) {
		if cli == nil {
			return
		}
		cli.closeLink(link.id)
	}
	close(link.endPoints[0])
	close(link.endPoints[1])
}

// Handler handler
type Handler struct {
	cfg         *global.Configure
	lockClients sync.RWMutex
	clients     map[string]*clients // client id => client
	lockLinks   sync.RWMutex
	links       map[string]link // link id => endpoints
}

// New create handler
func New(cfg *global.Configure) *Handler {
	return &Handler{
		cfg:     cfg,
		clients: make(map[string]*clients),
		links:   make(map[string]link),
	}
}

// Handle main loop
func (h *Handler) Handle(conn net.Conn) {
	c := network.NewConn(conn)
	var id string
	var idx uint32
	defer func() {
		if len(id) > 0 {
			logging.Info("%s disconnected", id)
		}
		c.Close()
	}()
	var err error
	for i := 0; i < 10; i++ {
		id, idx, err = h.readHandshake(c)
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
	logging.Info("%s-%d connected", id, idx)

	clients := h.tryGetClients(id)
	cli := clients.new(idx, c)

	defer h.closeClient(cli)
	go cli.keepalive()

	cli.run()
}

func (h *Handler) tryGetClients(id string) *clients {
	h.lockClients.Lock()
	defer h.lockClients.Unlock()
	clients := h.clients[id]
	if clients != nil {
		return clients
	}
	clients = newClients(h, id)
	h.clients[id] = clients
	return clients
}

// readHandshake read handshake message and compare secret encoded from md5
func (h *Handler) readHandshake(c *network.Conn) (string, uint32, error) {
	msg, _, err := c.ReadMessage(5 * time.Second)
	if err != nil {
		return "", 0, err
	}
	if msg.GetXType() != network.Msg_handshake {
		return "", 0, errNotHandshake
	}
	n := bytes.Compare(msg.GetHsp().GetEnc(), h.cfg.Enc[:])
	if n != 0 {
		return "", 0, errInvalidHandshake
	}
	return msg.GetFrom(), msg.GetFromIdx(), nil
}

func (h *Handler) getClient(linkID, to string, toIdx uint32) *client {
	h.lockLinks.RLock()
	link := h.links[linkID]
	h.lockLinks.RUnlock()

	if link.endPoints[0] != nil && link.endPoints[0].is(to, toIdx) {
		return link.endPoints[0]
	}
	if link.endPoints[1] != nil && link.endPoints[1].is(to, toIdx) {
		return link.endPoints[1]
	}

	h.lockClients.RLock()
	clients := h.clients[to]
	h.lockClients.RUnlock()

	if clients == nil {
		return nil
	}
	return clients.next()
}

func (h *Handler) onMessage(from *client, conn *network.Conn, msg *network.Msg, size uint16) {
	to := msg.GetTo()
	toIdx := msg.GetToIdx()
	if msg.GetXType() == network.Msg_keepalive {
		return
	}
	cli := h.getClient(msg.GetLinkId(), to, toIdx)
	if cli == nil {
		logging.Error("client %s-%d not found", to, toIdx)
		return
	}
	h.msgHook(msg, from, cli, size)
	err := cli.writeMessage(msg)
	if err != nil {
		logging.Error("write message %s from %s-%d to %s-%d: %v",
			msg.GetXType().String(),
			msg.GetFrom(), msg.GetFromIdx(),
			msg.GetTo(), msg.GetToIdx(),
			err)
	}
}

func (h *Handler) addLink(name, id string, t network.ConnectRequestType, from, to *client) {
	var link link
	link.id = id
	link.t = t
	if from != nil {
		from.addLink(id)
		link.endPoints[0] = from
	}
	if to != nil {
		to.addLink(id)
		link.endPoints[1] = to
	}
	h.lockLinks.Lock()
	h.links[id] = link
	h.lockLinks.Unlock()
	logging.Info("add link %s name %s from %s-%d to %s-%d",
		id, name, from.parent.id, from.idx, to.parent.id, to.idx)
}

func (h *Handler) removeLink(id string, from, to *client) {
	if from != nil {
		from.removeLink(id)
	}
	if to != nil {
		to.removeLink(id)
	}
	h.lockLinks.Lock()
	delete(h.links, id)
	h.lockLinks.Unlock()
	logging.Info("remove link %s from %s-%d to %s-%d",
		id, from.parent.id, from.idx, to.parent.id, to.idx)
}

func (h *Handler) responseLink(id string, ok bool, msg string, from, to *client) {
	if ok {
		logging.Info("link %s from %s-%d to %s-%d connect successed",
			id, from.parent.id, from.idx, to.parent.id, to.idx)
	} else {
		logging.Info("link %s from %s-%d to %s-%d connect failed, %s",
			id, from.parent.id, from.idx, to.parent.id, to.idx, msg)
		// TODO: remove link?
	}
}

// msgHook hook from on message
func (h *Handler) msgHook(msg *network.Msg, from, to *client, size uint16) {
	switch msg.GetXType() {
	// create link
	case network.Msg_connect_req:
		h.addLink(msg.GetCreq().GetName(), msg.GetLinkId(), msg.GetCreq().GetXType(), from, to)
	// remove link
	case network.Msg_disconnect:
		h.removeLink(msg.GetLinkId(), from, to)
	// response link
	case network.Msg_connect_rep:
		rep := msg.GetCrep()
		h.responseLink(msg.GetLinkId(), rep.GetOk(), rep.GetMsg(), from, to)
	// forward data
	case network.Msg_forward:
		data := msg.GetXData()
		logging.Debug("link %s forward %d bytes from %s-%d to %s-%d",
			msg.GetLinkId(), len(data.GetData()), from.parent.id, from.idx, to.parent.id, to.idx)
	case network.Msg_shell_data:
		data := msg.GetSdata()
		logging.Debug("shell %s forward %d bytes from %s-%d to %s-%d",
			msg.GetLinkId(), len(data.GetData()), from.parent.id, from.idx, to.parent.id, to.idx)
	// shell
	case network.Msg_shell_resize:
		data := msg.GetSresize()
		logging.Info("shell %s from %s-%d to %s-%d resize to (%d,%d)",
			msg.GetLinkId(), from.parent.id, from.idx, to.parent.id, to.idx,
			data.GetRows(), data.GetCols())
	}
	msg.From = from.parent.id
	msg.FromIdx = from.idx
	msg.To = to.parent.id
	msg.ToIdx = to.idx
	logging.Debug("forward %d bytes on link %s from %s-%d to %s-%d", size, msg.GetLinkId(),
		from.parent.id, from.idx, to.parent.id, to.idx)
}

func (h *Handler) closeClient(cli *client) {
	links := cli.getLinks()
	for _, t := range links {
		h.lockLinks.RLock()
		link := h.links[t]
		h.lockLinks.RUnlock()
		link.close()
		h.lockLinks.Lock()
		delete(h.links, t)
		h.lockLinks.Unlock()
	}
	h.lockClients.RLock()
	clients := h.clients[cli.parent.id]
	h.lockClients.RUnlock()
	if clients != nil {
		clients.close(cli.idx)
	}
}

func (h *Handler) removeClients(id string) {
	h.lockClients.Lock()
	delete(h.clients, id)
	h.lockClients.Unlock()
}
