//go:build !windows
// +build !windows

package core

import (
	"image"
	"time"
)

// CreateWorkerProcess create worker process
func CreateWorkerProcess(confDir string) (*Process, error) {
	return nil, ErrNotSupported
}

// Close close process
func (p *Process) Close() {
}

func (p *Process) Capture(timeout time.Duration) (*image.RGBA, error) {
	return nil, nil
}
