//go:build !vnc
// +build !vnc

package worker

import "natpass/code/client/tunnel/vnc/vncnetwork"

func runMouse(data *vncnetwork.MouseData) {
}

func runKeyboard(data *vncnetwork.KeyboardData) {
}

func runScroll(data *vncnetwork.ScrollData) {
}
