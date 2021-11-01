//go:build !windows
// +build !windows

package worker

import "github.com/gorilla/websocket"

func (worker *Worker) runCapture(conn *websocket.Conn) {
}
