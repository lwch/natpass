package handler

import (
	"context"
	"natpass/code/client/global"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"net"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type Handler struct {
	sync.RWMutex
	ctx context.Context
	ctl network.Natpass_ControlClient
	fwd network.Natpass_ForwardClient
	mgr *tunnel.Mgr

	// runtime
	conns map[string]net.Conn
}

func New(ctx context.Context,
	ctl network.Natpass_ControlClient, fwd network.Natpass_ForwardClient) *Handler {
	return &Handler{
		ctx: ctx,
		ctl: ctl,
		fwd: fwd,
		mgr: tunnel.NewMgr(),
	}
}

func (h *Handler) CreateTunnels(id string, tunnels []global.Tunnel) {
	cfgs := make(map[string]global.Tunnel, len(tunnels)) // name => config
	h.conns = make(map[string]net.Conn)                  // local_cid => connection
	for _, t := range tunnels {
		l, err := tunnel.NewListener(t.Type, t.Name, t.LocalAddr, t.LocalPort)
		runtime.Assert(err)
		cfgs[t.Name] = t
		go l.Loop(func(c net.Conn, name string) {
			id := createTunnel(h.ctx, cfgs[name], h.ctl, id)
			h.Lock()
			h.conns[id] = c
			h.Unlock()
		})
	}
}

func (h *Handler) HandleControl() {
	for {
		data, err := h.ctl.Recv()
		if err != nil {
			logging.Error("handler: %v", err)
			return
		}
		switch data.GetXType() {
		case network.ControlData_connect_req:
			h.handleConnectReq(data)
		case network.ControlData_connect_rep:
			h.handleConnectRep(data)
		}
	}
}

func (h *Handler) handleConnectReq(data *network.ControlData) {
	req := data.GetCreq()
	cid, err := runtime.UUID(cidLength)
	rep := &network.ConnectResponse{
		Name:      req.GetName(),
		RemoteCid: req.GetCid(),
	}
	if err != nil {
		logging.Error("generate channel_id failed, err=%v", err)
		rep.Ok = false
		rep.Msg = err.Error()
		connectResponse(h.ctl, rep, data.GetTo(), data.GetFrom())
		return
	}
	t := "tcp"
	if req.GetXType() == network.ConnectRequest_udp {
		t = "udp"
	}
	tn, err := tunnel.NewConnect(req.GetName(), data.GetTo(), cid,
		data.GetFrom(), req.GetCid(),
		t, req.GetAddr(), req.GetPort())
	if err != nil {
		logging.Error("new connect failed: %v", err)
		rep.Ok = false
		rep.Msg = err.Error()
		connectResponse(h.ctl, rep, data.GetTo(), data.GetFrom())
		return
	}
	h.mgr.Add(tn)
	rep.Ok = true
	rep.LocalCid = cid
	connectResponse(h.ctl, rep, data.GetTo(), data.GetFrom())
}

func (h *Handler) handleConnectRep(data *network.ControlData) {
	rep := data.GetCrep()
	if !rep.GetOk() {
		logging.Error("connect %s failed, err=%s", rep.GetName(), rep.GetMsg())
		return
	}
	tn, err := tunnel.NewListen(rep.GetName(), data.GetTo(), rep.GetRemoteCid(),
		data.GetFrom(), rep.GetLocalCid(), h.conns[rep.GetRemoteCid()])
	runtime.Assert(err)
	h.mgr.Add(tn)
	go tn.ForwardLocal(h.fwd)
}

func (h *Handler) HandleForward() {
	for {
		data, err := h.fwd.Recv()
		if err != nil {
			logging.Error("read data from remote failed, err=%v", err)
			continue
		}
		tn := h.mgr.Find(data.GetCid())
		if tn == nil {
			logging.Error("tunnel not found, from=%s, to=%s", data.GetFrom(), data.GetTo())
			continue
		}
		tn.WriteLocal(data.GetData())
	}
}
