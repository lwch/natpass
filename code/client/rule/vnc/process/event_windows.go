package process

import (
	"natpass/code/client/tunnel/vnc/define"
	"syscall"

	"github.com/lwch/logging"
)

// CADEvent handle ctrl+alt+del event
func (p *Process) CADEvent() {
	ok, _, err := syscall.Syscall(define.FuncSendSAS, 1, 0, 0, 0)
	if ok == 0 {
		logging.Error("send sas failed: %v", err)
	}
}
