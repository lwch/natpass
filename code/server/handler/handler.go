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
	clients map[string]*client
}

func New(cfg *global.Configure) *Handler {
	return &Handler{
		cfg:     cfg,
		clients: make(map[string]*client),
	}
}

func (h *Handler) Handle(conn net.Conn) {
	c := network.NewConn(conn)
	defer c.Close()
	var id string
	var err error
	for i := 0; i < 10; i++ {
		id, err = h.readHandshake(c)
		if err != nil {
			if err == errInvalidHandshake {
				logging.Error("invalid handshake from %s", c.RemoteAddr().String())
				return
			}
			logging.Error("read handshake from %s %d times, err=%v", c.RemoteAddr().String(), err)
			continue
		}
		break
	}
	if err != nil {
		return
	}
	logging.Info("%s connected", id)

	cli := newClient(id, c)
	h.Lock()
	h.clients[cli.id] = cli
	h.Unlock()

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
