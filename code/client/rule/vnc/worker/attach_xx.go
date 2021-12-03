//go:build !windows
// +build !windows

package worker

func attachDesktop() (func(), error) {
	return func() {}, nil
}
