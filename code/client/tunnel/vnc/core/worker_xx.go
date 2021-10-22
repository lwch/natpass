// +build !windows

package core

type workerOsBased struct {
}

func (worker *Worker) init() error {
	return nil
}
