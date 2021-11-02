package vnc

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"natpass/code/client/pool"
	"natpass/code/network"
	"natpass/code/utils"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

var upgrader = websocket.Upgrader{}

func (v *VNC) WS(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ws/")
	conn := pool.Get(id)
	if conn == nil {
		http.NotFound(w, r)
		return
	}
	local, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ch := conn.ChanRead(id)
	defer conn.SendDisconnect(v.link.target, v.link.targetIdx, v.link.id)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer cancel()
		defer wg.Done()
		v.remoteRead(ctx, ch, local)
	}()
	go func() {
		defer cancel()
		defer wg.Done()
		v.localRead(ctx, local, conn)
	}()
	wg.Wait()
}

func (v *VNC) remoteRead(ctx context.Context, ch <-chan *network.Msg, local *websocket.Conn) {
	defer utils.Recover("remoteRead")
	for {
		var msg *network.Msg
		select {
		case msg = <-ch:
		case <-ctx.Done():
			return
		}
		switch msg.GetXType() {
		case network.Msg_vnc_image:
			data, err := decodeImage(msg.GetVimg())
			runtime.Assert(err)
			replyImage(local, msg.GetVimg(), data)
		default:
			logging.Error("on message: %s", msg.GetXType().String())
			return
		}
	}
}

func (v *VNC) localRead(ctx context.Context, local *websocket.Conn, remote *pool.Conn) {
	defer utils.Recover("localRead")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, data, err := local.ReadMessage()
		if err != nil {
			logging.Error("local read: %v", err)
			return
		}
		var msg struct {
			Action string `json:"action"`
		}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			logging.Error("unmarshal: %v", err)
			return
		}
		switch msg.Action {
		case "mouse":
			v.mouseEvent(remote, data)
		case "keyboard":
		case "cad":
		}
	}
}

func decodeImage(data *network.VncImage) ([]byte, error) {
	switch data.GetEncode() {
	case network.VncImage_raw:
		return data.GetData(), nil
	case network.VncImage_jpeg:
		img, err := jpeg.Decode(bytes.NewReader(data.GetData()))
		if err != nil {
			return nil, err
		}
		// dumpImage(img)
		rect := img.Bounds()
		raw := image.NewRGBA(rect)
		draw.Draw(raw, rect, img, rect.Min, draw.Src)
		return raw.Pix, nil
	case network.VncImage_png:
		// TODO
	}
	return nil, errors.New("unsupported")
}

func dumpImage(img image.Image) {
	f, err := os.Create(`./debug.jpeg`)
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

func replyImage(conn *websocket.Conn, msg *network.VncImage, data []byte) {
	info := msg.GetXInfo()
	buf := make([]byte, len(data)+24)
	binary.BigEndian.PutUint32(buf, info.GetScreenWidth())
	binary.BigEndian.PutUint32(buf[4:], info.GetScreenHeight())
	binary.BigEndian.PutUint32(buf[8:], info.GetRectX())
	binary.BigEndian.PutUint32(buf[12:], info.GetRectY())
	binary.BigEndian.PutUint32(buf[16:], info.GetRectWidth())
	binary.BigEndian.PutUint32(buf[20:], info.GetRectHeight())
	copy(buf[24:], data)
	conn.WriteMessage(websocket.BinaryMessage, buf)
}
