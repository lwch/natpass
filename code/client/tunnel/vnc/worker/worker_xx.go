//go:build !windows
// +build !windows

package worker

type workerOsBased struct {
}

func (worker *Worker) init() error {
	return nil
}
