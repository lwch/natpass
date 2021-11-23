//go:build !vnc
// +build !vnc

package worker

import (
	"natpass/code/client/rule/vnc/vncnetwork"

	"github.com/gorilla/websocket"
)

func runMouse(data *vncnetwork.MouseData) {
}

func runKeyboard(data *vncnetwork.KeyboardData) {
}

func runScroll(data *vncnetwork.ScrollData) {
}

func runClipboard(conn *websocket.Conn, data *vncnetwork.ClipboardData) {
}
