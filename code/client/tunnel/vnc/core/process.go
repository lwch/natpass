package core

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

const (
	listenBegin = 6155
	listenEnd   = 6955
)

func (p *Process) listenAndServe() (uint16, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ws)
	port := uint16(listenBegin)
	for {
		if port > listenEnd {
			return 0, errors.New("no port available")
		}
		p.srv = &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", port),
			Handler: mux,
		}
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			port++
			continue
		}
		go p.srv.Serve(ln)
		return port, nil
	}
}

var upgrader = websocket.Upgrader{EnableCompression: true}

func (p *Process) ws(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
}

func (p *Process) kill() {
	ps, _ := os.FindProcess(p.pid)
	if ps != nil {
		ps.Kill()
	}
}
