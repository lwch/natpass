//go:build !windows
// +build !windows

package process

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/lwch/natpass/code/client/rule/vnc/vncnetwork"
)

// CreateWorker create worker process
func CreateWorker(name, confDir string, showCursor bool) (*Process, error) {
	dir, err := os.Executable()
	if err != nil {
		return nil, err
	}
	var p Process
	p.chWrite = make(chan *vncnetwork.VncMsg)
	p.chImage = make(chan *vncnetwork.ImageData)
	p.chClipboard = make(chan *vncnetwork.ClipboardData)
	port, err := p.listenAndServe()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(dir, "vnc", "--conf", confDir,
		"--name", name,
		"--port", fmt.Sprintf("%d", port))
	err = cmd.Start()
	if err != nil {
		p.Close()
		return nil, err
	}
	return &p, nil
}
