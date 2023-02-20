package handler

import (
	"bytes"
	"net"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/natpass/code/server/global"
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
		cli.sendClose(link.id)
	}
	close(link.endPoints[0])
	close(link.endPoints[1])
}

// Handler handler
type Handler struct {
	cfg       *global.Configure
	clis      *clients
	lockLinks sync.RWMutex
	links     map[string]link // link id => endpoints
}

// New create handler
func New(cfg *global.Configure) *Handler {
	h := &Handler{
		cfg:   cfg,
		links: make(map[string]link),
	}
	h.clis = newClients(h)
	return h
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

	cli := h.clis.new(id, c)

	defer h.clis.close(id)
	go cli.keepalive()

	cli.run()
}

// readHandshake read handshake message and compare secret encoded from md5
func (h *Handler) readHandshake(c *network.Conn) (string, error) {
	msg, _, err := c.ReadMessage(5 * time.Second)
	if err != nil {
		return "", err
	}
	if msg.GetXType() != network.Msg_handshake {
		return "", errNotHandshake
	}
	n := bytes.Compare(msg.GetHsp().GetEnc(), h.cfg.Hasher.Hash())
	if n != 0 {
		return "", errInvalidHandshake
	}
	return msg.GetFrom(), nil
}

func (h *Handler) getClient(linkID, to string) *client {
	h.lockLinks.RLock()
	link := h.links[linkID]
	h.lockLinks.RUnlock()

	if link.endPoints[0] != nil && link.endPoints[0].id == to {
		return link.endPoints[0]
	}
	if link.endPoints[1] != nil && link.endPoints[1].id == to {
		return link.endPoints[1]
	}

	return h.clis.lookup(to)
}

func (h *Handler) onMessage(from *client, conn *network.Conn, msg *network.Msg, size uint16) {
	to := msg.GetTo()
	if msg.GetXType() == network.Msg_keepalive {
		return
	}
	cli := h.getClient(msg.GetLinkId(), to)
	if cli == nil {
		logging.Error("client %s not found", to)
		return
	}
	h.msgHook(msg, from, cli, size)
	err := cli.writeMessage(msg)
	if err != nil {
		logging.Error("write message %s from %s to %s: %v",
			msg.GetXType().String(),
			msg.GetFrom(), msg.GetTo(),
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
	logging.Info("add link %s name %s from %s to %s",
		id, name, from.id, to.id)
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
	logging.Info("remove link %s from %s to %s",
		id, from.id, to.id)
}

func (h *Handler) responseLink(id string, ok bool, msg string, from, to *client) {
	if ok {
		logging.Info("link %s from %s to %s connect successed",
			id, from.id, to.id)
	} else {
		logging.Info("link %s from %s to %s connect failed, %s",
			id, from.id, to.id, msg)
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
		logging.Debug("link %s forward %d bytes from %s to %s",
			msg.GetLinkId(), len(data.GetData()), from.id, to.id)
	case network.Msg_shell_data:
		data := msg.GetSdata()
		logging.Debug("shell %s forward %d bytes from %s to %s",
			msg.GetLinkId(), len(data.GetData()), from.id, to.id)
	// shell
	case network.Msg_shell_resize:
		data := msg.GetSresize()
		logging.Info("shell %s from %s to %s resize to (%d,%d)",
			msg.GetLinkId(), from.id, to.id,
			data.GetRows(), data.GetCols())
	}
	msg.From = from.id
	msg.To = to.id
	logging.Debug("forward %d bytes on link %s from %s to %s", size, msg.GetLinkId(),
		from.id, to.id)
}

func (h *Handler) closeLink(id string) {
	h.lockLinks.RLock()
	link := h.links[id]
	h.lockLinks.RUnlock()
	link.close()
	h.lockLinks.Lock()
	delete(h.links, id)
	h.lockLinks.Unlock()
}
