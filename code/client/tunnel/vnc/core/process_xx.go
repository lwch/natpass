// +build !windows

package core

// Process process
type Process struct {
}

// CreateWorkerProcess create worker process
func CreateWorkerProcess() (*Process, error) {
	return nil, ErrNotSupported
}

// Close close process
func (p *Process) Close() {
}
