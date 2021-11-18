package shell

import (
	"io"
	"natpass/code/client/pool"
	"natpass/code/network"
	"natpass/code/utils"
	"os"

	"github.com/lwch/logging"
	"golang.org/x/text/encoding/simplifiedchinese"
	"google.golang.org/protobuf/proto"
)

// Link shell link
type Link struct {
	parent    *Shell
	id        string // link id
	target    string // target id
	targetIdx uint32 // target idx
	remote    *pool.Conn
	// in remote
	pid    int
	stdin  io.WriteCloser
	stdout io.ReadCloser
	// runtime
	sendBytes  uint64
	recvBytes  uint64
	sendPacket uint64
	recvPacket uint64
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

// Close close link
func (link *Link) Close() {
	link.onClose()
	p, err := os.FindProcess(link.pid)
	if err == nil {
		p.Kill()
	}
	link.remote.SendDisconnect(link.target, link.targetIdx, link.id)
	link.parent.remove(link.id)
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
		data, _ := proto.Marshal(msg)
		link.recvBytes += uint64(len(data))
		link.recvPacket++
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
		case network.Msg_disconnect:
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
		var data []byte
		switch {
		case isUtf8(buf[:n]):
			data = buf[:n]
		case isGBK(buf[:n]):
			data, err = simplifiedchinese.GBK.NewDecoder().Bytes(buf[:n])
			if err != nil {
				logging.Error("transform gbk to utf8 failed: %v", err)
				continue
			}
		}
		logging.Debug("link %s on shell %s read from local %d bytes",
			link.id, link.parent.Name, n)
		send := link.remote.SendShellData(link.target, link.targetIdx, link.id, data)
		link.sendBytes += send
		link.sendPacket++
	}
}

// SendData send data
func (link *Link) SendData(data []byte) {
	send := link.remote.SendShellData(link.target, link.targetIdx, link.id, data)
	link.sendBytes += send
	link.sendPacket++
}

// SendResize send resize message
func (link *Link) SendResize(rows, cols uint32) {
	link.remote.SendShellResize(link.target, link.targetIdx, link.id, rows, cols)
}
