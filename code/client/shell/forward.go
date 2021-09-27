package shell

import (
	"natpass/code/client/pool"
	"natpass/code/network"
	"natpass/code/utils"

	"github.com/lwch/logging"
)

func (shell *Shell) Forward(remote *pool.Conn, toIdx uint32) {
	go shell.remoteRead(remote)
	go shell.localRead(remote, toIdx)
}

func (shell *Shell) remoteRead(remote *pool.Conn) {
	defer utils.Recover("remoteRead")
	defer shell.Close()
	ch := remote.ChanRead(shell.id)
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		switch msg.GetXType() {
		case network.Msg_shell_resize:
			// TODO
		case network.Msg_shell_data:
			// TODO
		case network.Msg_shell_close:
			// TODO
		}
	}
}

func (shell *Shell) localRead(remote *pool.Conn, toIdx uint32) {
	defer utils.Recover("localRead")
	defer shell.Close()
	buf := make([]byte, 16*1024)
	for {
		n, err := shell.stdout.Read(buf)
		if err != nil {
			// if !link.closeFromRemote {
			// 	link.remote.SendDisconnect(link.target, link.targetIdx, link.id)
			// }
			logging.Error("read data on shell %s link %s failed, err=%v", shell.Name, shell.id, err)
			return
		}
		if n == 0 {
			continue
		}
		logging.Debug("link %s on shell %s read from local %d bytes", shell.id, shell.Name, n)
		remote.SendShellData(shell.cfg.Target, toIdx, shell.id, buf[:n])
	}
}
