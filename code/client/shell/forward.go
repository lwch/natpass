package shell

import (
	"natpass/code/network"
	"natpass/code/utils"

	"github.com/lwch/logging"
)

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
		case network.Msg_shell_created:
			if msg.GetCrep().GetOk() {
				link.OnWork <- struct{}{}
				continue
			}
			logging.Error("create shell %s on tunnel %s failed, err=%s",
				link.id, link.parent.Name, msg.GetCrep().GetMsg())
			return
		case network.Msg_shell_resize:
			// TODO
		case network.Msg_shell_data:
			_, err := link.stdin.Write(msg.GetSdata().GetData())
			if err != nil {
				// TODO: close
				logging.Error("write data on shell %s link %s failed, err=%v",
					link.parent.Name, link.id, err)
				return
			}
		case network.Msg_shell_close:
			// TODO
		}
	}
}

func (link *Link) localRead() {
	defer utils.Recover("localRead")
	defer link.Close()
	<-link.OnWork
	buf := make([]byte, 16*1024)
	for {
		n, err := link.stdout.Read(buf)
		if err != nil {
			// if !link.closeFromRemote {
			// 	link.remote.SendDisconnect(link.target, link.targetIdx, link.id)
			// }
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
