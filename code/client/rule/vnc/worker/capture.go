package worker

import (
	"fmt"

	"github.com/jkstack/natpass/code/client/rule/vnc/vncnetwork"
	"github.com/lwch/logging"
)

func (worker *Worker) runCapture() vncnetwork.ImageData {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: " + err.Error())
		return vncnetwork.ImageData{
			Ok:  false,
			Msg: fmt.Sprintf("attach desktop: " + err.Error()),
		}
	}
	defer detach()
	img, err := worker.cli.Screenshot()
	if err != nil {
		logging.Error("screenshot: " + err.Error())
		return vncnetwork.ImageData{
			Ok:  false,
			Msg: fmt.Sprintf("screenshot: " + err.Error()),
		}
	}
	data := make([]byte, len(img.Pix))
	copy(data, img.Pix)
	return vncnetwork.ImageData{
		Ok:     true,
		Bits:   32,
		Width:  uint32(img.Rect.Max.X),
		Height: uint32(img.Rect.Max.Y),
		Data:   data,
	}
}
