package tunnel

import (
	"fmt"
	"net"

	"github.com/lwch/logging"
)

type Listener struct {
	name string
	l    net.Listener
}

func NewListener(t, name, addr string, port uint16) (*Listener, error) {
	l, err := net.Listen(t, fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}
	return &Listener{
		name: name,
		l:    l,
	}, nil
}

func (l *Listener) Close() {
	l.l.Close()
}

func (l *Listener) Loop(cb func(net.Conn, string)) {
	for {
		conn, err := l.l.Accept()
		if err != nil {
			logging.Error("accept: %v", err)
			continue
		}
		go cb(conn, l.name)
	}
}
