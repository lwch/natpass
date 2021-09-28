package tunnel

import (
	"bytes"
	"io"
	"natpass/code/client/pool"
	"natpass/code/network"
	"natpass/code/utils"
	"net"

	"github.com/lwch/logging"
)

type Link struct {
	parent          *Tunnel
	id              string // link id
	target          string // target id
	targetIdx       uint32 // target idx
	local           net.Conn
	remote          *pool.Conn
	OnWork          chan struct{}
	closeFromRemote bool
}

// NewLink create link
func NewLink(parent *Tunnel, id, target string, local net.Conn, remote *pool.Conn) *Link {
	remote.AddLink(id)
	logging.Info("create link %s for tunnel %s on connection %d",
		id, parent.Name, remote.Idx)
	return &Link{
		parent: parent,
		id:     id,
		target: target,
		local:  local,
		remote: remote,
		OnWork: make(chan struct{}),
	}
}

func (link *Link) close() {
	logging.Info("close link %s on tunnel %s", link.id, link.parent.Name)
	link.local.Close()
	link.remote.RemoveLink(link.id)
}

// Forward forward data
func (link *Link) Forward() {
	go link.remoteRead()
	go link.localRead()
}

func (link *Link) remoteRead() {
	defer utils.Recover("remoteRead")
	defer link.close()
	ch := link.remote.ChanRead(link.id)
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		link.targetIdx = msg.GetFromIdx()
		switch msg.GetXType() {
		case network.Msg_forward:
			_, err := io.Copy(link.local, bytes.NewReader(msg.GetXData().GetData()))
			if err != nil {
				logging.Error("write data on tunnel %s link %s failed, err=%v", link.parent.Name, link.id, err)
				return
			}
		case network.Msg_connect_rep:
			if msg.GetCrep().GetOk() {
				link.OnWork <- struct{}{}
				continue
			}
			logging.Error("create link %s on tunnel %s failed, err=%s",
				link.id, link.parent.Name, msg.GetCrep().GetMsg())
			return
		case network.Msg_disconnect:
			logging.Info("disconnect link %s on tunnel %s from remote",
				link.id, link.parent.Name)
			link.closeFromRemote = true
			return
		}
	}
}

func (link *Link) localRead() {
	defer utils.Recover("localRead")
	defer link.close()
	<-link.OnWork
	buf := make([]byte, 16*1024)
	for {
		n, err := link.local.Read(buf)
		if err != nil {
			if !link.closeFromRemote {
				link.remote.SendDisconnect(link.target, link.targetIdx, link.id)
			}
			logging.Error("read data on tunnel %s link %s failed, err=%v", link.parent.Name, link.id, err)
			return
		}
		if n == 0 {
			continue
		}
		logging.Debug("link %s on tunnel %s read from local %d bytes", link.id, link.parent.Name, n)
		link.remote.SendData(link.target, link.targetIdx, link.id, buf[:n])
	}
}

// SetTargetIdx set link remote index
func (link *Link) SetTargetIdx(idx uint32) {
	link.targetIdx = idx
}
