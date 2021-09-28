package shell

import (
	"io"
	"natpass/code/client/pool"
	"natpass/code/network"
	"natpass/code/utils"
	"os"

	"github.com/lwch/logging"
)

// Link shell link
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
	link.remote.SendShellClose(link.target, link.targetIdx, link.id)
}

// Forward forward data
func (link *Link) Forward() {
	go link.remoteRead()
	go link.localRead()
}

func (link *Link) remoteRead() {
	defer utils.Recover("remoteRead")
	defer link.Close()
	ch := link.remote.ChanRead(link.id)
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		link.targetIdx = msg.GetFromIdx()
		switch msg.GetXType() {
		case network.Msg_shell_resize:
			size := msg.GetSresize()
			link.resize(size.GetRows(), size.GetCols())
		case network.Msg_shell_data:
			_, err := link.stdin.Write(msg.GetSdata().GetData())
			if err != nil {
				logging.Error("write data on shell %s link %s failed, err=%v",
					link.parent.Name, link.id, err)
				return
			}
		case network.Msg_shell_close:
			logging.Info("shell %s link %s closed by remote", link.parent.Name, link.id)
			return
		}
	}
}

func (link *Link) localRead() {
	defer utils.Recover("localRead")
	defer link.Close()
	buf := make([]byte, 16*1024)
	for {
		n, err := link.stdout.Read(buf)
		if err != nil {
			logging.Error("read data on shell %s link %s failed, err=%v",
				link.parent.Name, link.id, err)
			return
		}
		if n == 0 {
			continue
		}
		logging.Debug("link %s on shell %s read from local %d bytes",
			link.id, link.parent.Name, n)
		link.remote.SendShellData(link.target, link.targetIdx, link.id, buf[:n])
	}
}

// SendData send data
func (link *Link) SendData(data []byte) {
	link.remote.SendShellData(link.target, link.targetIdx, link.id, data)
}

// SendResize send resize message
func (link *Link) SendResize(rows, cols uint32) {
	link.remote.SendShellResize(link.target, link.targetIdx, link.id, rows, cols)
}
