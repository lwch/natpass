package server

import (
	"context"
	"errors"
	"natpass/code/network"
	"natpass/code/server/global"
	"sync"

	"github.com/lwch/logging"
	"google.golang.org/grpc/metadata"
)

type Handler struct {
	sync.RWMutex
	cfg       *global.Configure
	chControl map[string]chan *network.ControlData
	chForward map[string]chan *network.Data
}

func NewHandler(cfg *global.Configure) *Handler {
	return &Handler{
		cfg:       cfg,
		chControl: make(map[string]chan *network.ControlData),
		chForward: make(map[string]chan *network.Data),
	}
}

func (h *Handler) Control(svr network.Natpass_ControlServer) error {
	md, ok := metadata.FromIncomingContext(svr.Context())
	if !ok {
		return errors.New("get context failed")
	}
	secret := md.Get("secret")
	if len(secret) == 0 {
		return errors.New("missing secret")
	}
	if secret[0] != h.cfg.Secret {
		return errors.New("invalid secret")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		data, err := svr.Recv()
		if err != nil {
			return err
		}
		switch data.GetXType() {
		case network.ControlData_come:
			id := data.GetFrom()
			logging.Info("client %s connected", id)
			defer h.close(id)
			h.Lock()
			h.chControl[id] = make(chan *network.ControlData, 10)
			h.chForward[id] = make(chan *network.Data, 10)
			h.Unlock()
			go h.forwardControl(ctx, svr, id)
		case network.ControlData_keepalive:
		default:
			h.RLock()
			ch := h.chControl[data.GetTo()]
			h.RUnlock()
			if ch == nil {
				logging.Error("forward control message failed, %s not connected", data.GetTo())
				continue
			}
			ch <- data
		}
	}
}

func (h *Handler) forwardControl(ctx context.Context, svr network.Natpass_ControlServer, id string) {
	h.RLock()
	ch := h.chControl[id]
	h.RUnlock()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-ch:
			svr.Send(data)
		}
	}
}

func (h *Handler) close(id string) {
	logging.Info("client %s disconnected", id)
	h.Lock()
	if ch, ok := h.chControl[id]; ok {
		close(ch)
		delete(h.chControl, id)
	}
	if ch, ok := h.chForward[id]; ok {
		close(ch)
		delete(h.chForward, id)
	}
	h.Unlock()
}

func (h *Handler) Forward(svr network.Natpass_ForwardServer) error {
	md, ok := metadata.FromIncomingContext(svr.Context())
	if !ok {
		return errors.New("get context failed")
	}
	secret := md.Get("secret")
	if len(secret) == 0 {
		return errors.New("missing secret")
	}
	if secret[0] != h.cfg.Secret {
		return errors.New("invalid secret")
	}
	id := md.Get("id")
	if len(id) == 0 {
		return errors.New("missing id")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.forward(ctx, svr, id[0])
	for {
		data, err := svr.Recv()
		if err != nil {
			return err
		}
		h.RLock()
		ch := h.chForward[data.GetTo()]
		h.RUnlock()
		if ch == nil {
			logging.Error("channel %s not found", data.GetTo())
			continue
		}
		ch <- data
	}
}

func (h *Handler) forward(ctx context.Context, svr network.Natpass_ForwardServer, id string) {
	h.RLock()
	ch := h.chForward[id]
	h.RUnlock()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-ch:
			svr.Send(data)
		}
	}
}
