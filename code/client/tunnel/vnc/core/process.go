package core

import (
	"errors"
	"fmt"
	"natpass/code/utils"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"google.golang.org/protobuf/proto"
)

const (
	listenBegin = 6155
	listenEnd   = 6955
)

// Process process
type Process struct {
	pid     int
	srv     *http.Server
	chWrite chan *VncMsg
	chImage chan *ImageData
}

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
	logging.Info("child process connected")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	defer p.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer utils.Recover("ws read")
		defer wg.Done()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				logging.Error("read message: %v", err)
				return
			}
			var msg VncMsg
			err = proto.Unmarshal(data, &msg)
			if err != nil {
				continue
			}
			switch msg.GetXType() {
			case VncMsg_capture_data:
				p.chImage <- msg.GetData()
			default:
			}
		}
	}()
	go func() {
		defer utils.Recover("ws write")
		defer wg.Done()
		for {
			msg := <-p.chWrite
			data, err := proto.Marshal(msg)
			if err != nil {
				continue
			}
			err = conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				logging.Error("write message: %v", err)
				return
			}
		}
	}()
	wg.Wait()
}

func (p *Process) kill() {
	ps, _ := os.FindProcess(p.pid)
	if ps != nil {
		ps.Kill()
	}
}
