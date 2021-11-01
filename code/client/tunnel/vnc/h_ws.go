package vnc

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"natpass/code/client/pool"
	"natpass/code/network"
	"net/http"
	"strings"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

func (v *VNC) WS(pool *pool.Pool, w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ws/")
	conn := pool.Get(id)
	if conn == nil {
		http.NotFound(w, r)
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
			logging.Info("%d", len(data))
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
		rect := img.Bounds()
		raw := image.NewRGBA(rect)
		draw.Draw(raw, rect, img, rect.Min, draw.Src)
		return raw.Pix, nil
	case network.VncImage_png:
		// TODO
	}
	return nil, errors.New("unsupported")
}
