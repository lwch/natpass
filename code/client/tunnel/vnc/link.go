package vnc

import (
	"natpass/code/client/pool"

	"github.com/lwch/logging"
)

// Link vnc link
type Link struct {
	parent    *VNC
	id        string // link id
	target    string // target id
	targetIdx uint32 // target idx
	remote    *pool.Conn
	// runtime
	sendBytes  uint64
	recvBytes  uint64
	sendPacket uint64
	recvPacket uint64
}

// NewLink create link
func NewLink(parent *VNC, id, target string, remote *pool.Conn) *Link {
	remote.AddLink(id)
	logging.Info("create vnc %s for tunnel %s on connection %d",
		id, parent.Name, remote.Idx)
	return &Link{
		parent: parent,
		id:     id,
		target: target,
		remote: remote,
	}
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

// SetTargetIdx set link remote index
func (link *Link) SetTargetIdx(idx uint32) {
	link.targetIdx = idx
}

// Fork fork worker process
func (link *Link) Fork() error {
	return nil
}

// Forward forward data
func (link *Link) Forward() {
}
