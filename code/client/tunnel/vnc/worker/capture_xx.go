//go:build !windows
// +build !windows

package worker

import "natpass/code/client/tunnel/vnc/vncnetwork"

func (worker *Worker) runCapture() vncnetwork.ImageData {
	return vncnetwork.ImageData{}
}
