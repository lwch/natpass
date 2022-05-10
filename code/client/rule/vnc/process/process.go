package process

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/rule/vnc/vncnetwork"
	"github.com/lwch/natpass/code/utils"
	"google.golang.org/protobuf/proto"
)

const (
	listenBegin = 6155
	listenEnd   = 6955
)

// Process process
type Process struct {
	pid         int
	srv         *http.Server
	chWrite     chan *vncnetwork.VncMsg
	chImage     chan *vncnetwork.ImageData
	chClipboard chan *vncnetwork.ClipboardData
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
			var msg vncnetwork.VncMsg
			err = proto.Unmarshal(data, &msg)
			if err != nil {
				continue
			}
			switch msg.GetXType() {
			case vncnetwork.VncMsg_capture_data:
				p.chImage <- msg.GetData()
			case vncnetwork.VncMsg_clipboard_event:
				p.chClipboard <- msg.GetClipboard()
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

// Close close process
func (p *Process) Close() {
	if p.srv != nil {
		p.srv.Close()
	}
	if p.chImage != nil {
		close(p.chImage)
		p.chImage = nil
	}
	if p.chClipboard != nil {
		close(p.chClipboard)
		p.chClipboard = nil
	}
	if p.chWrite != nil {
		close(p.chWrite)
		p.chWrite = nil
	}
	p.kill()
}

// Capture capture desktop image
func (p *Process) Capture(timeout time.Duration) (*image.RGBA, error) {
	var msg vncnetwork.VncMsg
	msg.XType = vncnetwork.VncMsg_capture_req
	p.chWrite <- &msg
	trans := func(data *vncnetwork.ImageData) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, int(data.GetWidth()), int(data.GetHeight())))
		copy(img.Pix, data.GetData())
		// dumpImage(img)
		return img
	}
	if timeout > 0 {
		select {
		case data := <-p.chImage:
			return trans(data), nil
		case <-time.After(timeout):
			return nil, errors.New("timeout")
		}
	} else {
		data := <-p.chImage
		return trans(data), nil
	}
}

func dumpImage(img image.Image) {
	f, err := os.Create(`C:\Users\lwch\Pictures\debug.jpeg`)
	if err != nil {
		logging.Error("debug: %v", err)
		return
	}
	defer f.Close()
	err = jpeg.Encode(f, img, nil)
	if err != nil {
		logging.Error("encode: %v", err)
		return
	}
}
