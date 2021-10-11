package utils

import "github.com/lwch/logging"

// Recover default recover
func Recover(name string) {
	if err := recover(); err != nil {
		logging.Error("%s: %v", name, err)
	}
}
