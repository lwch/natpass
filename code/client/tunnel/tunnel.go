package tunnel

import (
	"natpass/code/client/global"

	"github.com/lwch/logging"
)

type Tunnel struct {
	local  string
	remote string
}

func NewListen(name, local, remote string, cfg global.Tunnel) *Tunnel {
	return &Tunnel{
		local:  local,
		remote: remote,
	}
}

func NewConnect(name, local, remote string, t, addr string, port uint32) *Tunnel {
	logging.Info("new connect: name=%s, remote=%s://%s:%d", name, t, addr, port)
	return &Tunnel{
		local:  local,
		remote: remote,
	}
}
