// +build !windows

package shell

import (
	"errors"
	"os/exec"

	"github.com/creack/pty"
)

func (shell *Shell) Exec(id string) error {
	shell.id = id
	var cmd *exec.Cmd
	if len(shell.cfg.Exec) > 0 {
		cmd = exec.Command(shell.cfg.Exec)
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
	shell.stdin = f
	shell.stdout = f
	shell.pid = cmd.Process.Pid
	return nil
}

func (shell *Shell) onClose() {
	if shell.stdin != nil {
		shell.stdin.Close()
	}
}
