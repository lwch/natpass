//go:build !windows
// +build !windows

package worker

import "natpass/code/client/rule/vnc/vncnetwork"

func (worker *Worker) runCapture() vncnetwork.ImageData {
	return vncnetwork.ImageData{}
}
