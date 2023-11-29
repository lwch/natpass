package global

import (
	"net"

	"github.com/lwch/runtime"
)

func GeneratePort() uint16 {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{})
	runtime.Assert(err)
	defer l.Close()
	return uint16(l.Addr().(*net.TCPAddr).Port)
}
