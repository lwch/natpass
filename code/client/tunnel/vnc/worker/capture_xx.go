//go:build !windows
// +build !windows

package worker

import "github.com/gorilla/websocket"

func (worker *Worker) runCapture() vncnetwork.ImageData {
	return vncnetwork.ImageData{}
}
