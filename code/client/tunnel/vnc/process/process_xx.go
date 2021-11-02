//go:build !windows
// +build !windows

package process

import (
	"image"
	"time"
)

// CreateWorker create worker process
func CreateWorker(confDir string, showCursor bool) (*Process, error) {
	return nil, ErrNotSupported
}

// Close close process
func (p *Process) Close() {
}

func (p *Process) Capture(timeout time.Duration) (*image.RGBA, error) {
	return nil, nil
}
