package pool

import (
	"image"
	"natpass/code/network"
	"time"
)

// SendVNCImage send vnc image data
func (conn *Conn) SendVNCImage(to string, toIdx uint32, id string, screen, rect image.Rectangle,
	encode network.VncImageEncoding, data []byte) {
	var msg network.Msg
	msg.To = to
	msg.ToIdx = toIdx
	msg.XType = network.Msg_vnc_image
	msg.LinkId = id
	msg.Payload = &network.Msg_Vimg{
		Vimg: &network.VncImage{
			XInfo: &network.VncImageInfo{
				ScreenWidth:  uint32(screen.Dx()),
				ScreenHeight: uint32(screen.Dy()),
				RectX:        uint32(rect.Min.X),
				RectY:        uint32(rect.Min.Y),
				RectWidth:    uint32(rect.Dx()),
				RectHeight:   uint32(rect.Dy()),
			},
			Encode: encode,
			Data:   data,
		},
	}
	select {
	case conn.write <- &msg:
	case <-time.After(conn.parent.cfg.WriteTimeout):
	}
}
