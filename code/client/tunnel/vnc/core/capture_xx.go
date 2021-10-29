//go:build !windows

package core

import "github.com/gorilla/websocket"

func (worker *Worker) runCapture(conn *websocket.Conn) {
}
