package shell

import (
	"errors"
	"os"
	"os/exec"
)

// Exec execute shell command
func (link *Link) Exec() error {
	var cmd *exec.Cmd
	if len(link.parent.cfg.Exec) > 0 {
		cmd = exec.Command(link.parent.cfg.Exec)
	}
	if cmd == nil {
		dir, err := exec.LookPath("powershell")
		if err == nil {
			cmd = exec.Command(dir)
		}
	}
	if cmd == nil {
		dir, err := exec.LookPath("cmd")
		if err == nil {
			cmd = exec.Command(dir)
		}
	}
	if cmd == nil {
		return errors.New("no shell command supported")
	}
	cmd.Env = append(os.Environ(), link.parent.cfg.Env...)
	var err error
	link.stdin, err = cmd.StdinPipe()
	if err != nil {
		return err
	}
	link.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	go cmd.Wait() // defunct process
	link.pid = cmd.Process.Pid
	return nil
}

func (link *Link) onClose() {
	if link.stdin != nil {
		link.stdin.Close()
	}
	if link.stdout != nil {
		link.stdout.Close()
	}
}

func (link *Link) resize(rows, cols uint32) {
}
