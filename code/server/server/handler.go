package server

import "natpass/code/network"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Control(svr network.Natpass_ControlServer) error {
	return nil
}

func (h *Handler) Forward(svr network.Natpass_ForwardServer) error {
	return nil
}
