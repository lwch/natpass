// +build !windows

package shell

import (
	"errors"
	"os/exec"

	"github.com/creack/pty"
)

func (link *Link) Exec() error {
	var cmd *exec.Cmd
	if len(link.parent.cfg.Exec) > 0 {
		cmd = exec.Command(link.parent.cfg.Exec)
	}
	if cmd == nil {
		dir, err := exec.LookPath("bash")
		if err == nil {
			cmd = exec.Command(dir)
		}
	}
	if cmd == nil {
		dir, err := exec.LookPath("sh")
		if err == nil {
			cmd = exec.Command(dir)
		}
	}
	if cmd == nil {
		return errors.New("no shell command supported")
	}
	f, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	link.stdin = f
	link.stdout = f
	link.pid = cmd.Process.Pid
	return nil
}

func (link *Link) onClose() {
	if link.stdin != nil {
		link.stdin.Close()
	}
}
