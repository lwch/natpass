package shell

import (
	"io"
	"natpass/code/client/pool"
	"os"

	"github.com/lwch/logging"
)

type Link struct {
	parent    *Shell
	id        string // link id
	target    string // target id
	targetIdx uint32 // target idx
	remote    *pool.Conn
	onWork    chan struct{}
	// in remote
	pid    int
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

// NewLink create link
func NewLink(parent *Shell, id, target string, remote *pool.Conn) *Link {
	remote.AddLink(id)
	logging.Info("create shell %s for tunnel %s on connection %d",
		id, parent.Name, remote.Idx)
	return &Link{
		parent: parent,
		id:     id,
		target: target,
		remote: remote,
		onWork: make(chan struct{}),
	}
}

// SetTargetIdx set link remote index
func (link *Link) SetTargetIdx(idx uint32) {
	link.targetIdx = idx
}

// Close close link
func (link *Link) Close() {
	link.onClose()
	p, err := os.FindProcess(link.pid)
	if err == nil {
		p.Kill()
	}
}
