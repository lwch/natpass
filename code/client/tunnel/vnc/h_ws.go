package vnc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"natpass/code/client/pool"
	"natpass/code/network"
	"net/http"
	"os"
	"strings"

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
	for {
		msg := <-ch
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
