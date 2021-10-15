package reverse

import (
	"bytes"
	"io"
	"natpass/code/client/pool"
	"natpass/code/network"
	"natpass/code/utils"
	"net"

	"github.com/lwch/logging"
	"google.golang.org/protobuf/proto"
)

// Link link object
type Link struct {
	parent          *Tunnel
	id              string // link id
	target          string // target id
	targetIdx       uint32 // target idx
	local           net.Conn
	remote          *pool.Conn
	OnWork          chan struct{}
	closeFromRemote bool
	// runtime
	recvBytes  uint64
	sendBytes  uint64
	recvPacket uint64
	sendPacket uint64
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
	link.parent.remove(link.id)
}

// GetID get link id
func (link *Link) GetID() string {
	return link.id
}

// GetBytes get send and recv bytes
func (link *Link) GetBytes() (uint64, uint64) {
	return link.recvBytes, link.sendBytes
}

// GetPackets get send and recv packets
func (link *Link) GetPackets() (uint64, uint64) {
	return link.recvPacket, link.sendPacket
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
		data, _ := proto.Marshal(msg)
		link.recvBytes += uint64(len(data))
		link.recvPacket++
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
				n := link.remote.SendDisconnect(link.target, link.targetIdx, link.id)
				link.sendBytes += n
				link.sendPacket++
			}
			logging.Error("read data on tunnel %s link %s failed, err=%v", link.parent.Name, link.id, err)
			return
		}
		if n == 0 {
			continue
		}
		logging.Debug("link %s on tunnel %s read from local %d bytes", link.id, link.parent.Name, n)
		send := link.remote.SendData(link.target, link.targetIdx, link.id, buf[:n])
		link.sendBytes += send
		link.sendPacket++
	}
}

// SetTargetIdx set link remote index
func (link *Link) SetTargetIdx(idx uint32) {
	link.targetIdx = idx
}
