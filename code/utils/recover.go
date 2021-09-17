package utils

import "github.com/lwch/logging"

func Recover(name string) {
	if err := recover(); err != nil {
		logging.Error("%s: %v", name, err)
	}
}
