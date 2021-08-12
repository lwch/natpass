package main

import (
	"natpass/code/server/global"
	"natpass/code/server/handler"
	"net"
)

func run(cfg *global.Configure, l net.Listener) {
	h := handler.New(cfg)
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go h.Handle(conn)
	}
}
