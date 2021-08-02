package server

import (
	"errors"
	"natpass/code/network"
	"natpass/code/server/global"

	"github.com/lwch/logging"
	"google.golang.org/grpc/metadata"
)

type Handler struct {
	cfg *global.Configure
}

func NewHandler(cfg *global.Configure) *Handler {
	return &Handler{
		cfg: cfg,
	}
}

func (h *Handler) Control(svr network.Natpass_ControlServer) error {
	md, ok := metadata.FromIncomingContext(svr.Context())
	if !ok {
		return errors.New("get context failed")
	}
	id := md.Get("id")
	if len(id) == 0 {
		return errors.New("missing id")
	}
	secret := md.Get("secret")
	if len(secret) == 0 {
		return errors.New("missing secret")
	}
	if secret[0] != h.cfg.Secret {
		return errors.New("invalid secret")
	}
	defer func() {
		logging.Info("client %s disconnected", id[0])
	}()
	for {
		data, err := svr.Recv()
		if err != nil {
			return err
		}
		switch data.GetXType() {
		case network.ControlData_come:
			logging.Info("client %s connected", id[0])
		}
	}
}

func (h *Handler) Forward(svr network.Natpass_ForwardServer) error {
	return nil
}
