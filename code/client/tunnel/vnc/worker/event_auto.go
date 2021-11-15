//go:build vnc
// +build vnc

package worker

import (
	"natpass/code/client/tunnel/vnc/vncnetwork"

	"github.com/go-vgo/robotgo"
	"github.com/lwch/logging"
)

func runMouse(data *vncnetwork.MouseData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	robotgo.MoveMouse(int(data.GetX()), int(data.GetY()))
	var key string
	switch data.GetBtn() {
	case vncnetwork.MouseData_left:
		key = "left"
	case vncnetwork.MouseData_right:
		key = "right"
	case vncnetwork.MouseData_middle:
		key = "center"
	}
	switch data.GetType() {
	case vncnetwork.Status_down:
		robotgo.MouseToggle("down", key)
	case vncnetwork.Status_up:
		robotgo.MouseToggle("up", key)
	}
}

func runKeyboard(data *vncnetwork.KeyboardData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	switch data.Type {
	case vncnetwork.Status_down:
		robotgo.KeyToggle(data.Key, "down")
	case vncnetwork.Status_up:
		robotgo.KeyToggle(data.Key, "up")
	}
}

func runScroll(data *vncnetwork.ScrollData) {
	detach, err := attachDesktop()
	if err != nil {
		logging.Error("attach desktop: %v", err)
		return
	}
	defer detach()
	robotgo.Scroll(int(data.X), int(data.Y))
}
