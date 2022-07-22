package code

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
)

// Workspace workspace of code-server
type Workspace struct {
	parent *Code
	id     string
	target string
	name   string
	exec   *exec.Cmd
	remote *conn.Conn
	// runtime
	sendBytes  uint64
	recvBytes  uint64
	sendPacket uint64
	recvPacket uint64
}

func newWorkspace(parent *Code, id, name, target string, remote *conn.Conn) *Workspace {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return &Workspace{
		parent: parent,
		id:     id,
		target: target,
		name:   name,
		remote: remote,
	}
}

// GetID get link id
func (ws *Workspace) GetID() string {
	return ws.id
}

// GetBytes get send and recv bytes
func (ws *Workspace) GetBytes() (uint64, uint64) {
	return ws.recvBytes, ws.sendBytes
}

// GetPackets get send and recv packets
func (ws *Workspace) GetPackets() (uint64, uint64) {
	return ws.recvPacket, ws.sendPacket
}

// Exec execute code-server
func (ws *Workspace) Exec(dir string) error {
	workdir := filepath.Join(dir, ws.name)
	err := os.MkdirAll(workdir, 0755)
	if err != nil {
		logging.Error("can not create work dir[%s]: %v", workdir, err)
		return err
	}
	ws.exec = exec.Command("code-server", "--disable-update-check",
		"--auth", "none",
		"--socket", filepath.Join(workdir, ws.id+".sock"),
		"--user-data-dir", filepath.Join(workdir, "data"),
		"--extensions-dir", filepath.Join(workdir, "extensions"))
	stdout, err := ws.exec.StdoutPipe()
	if err != nil {
		logging.Error("can not get stdout pipe for link [%s] name [%s]", ws.id, ws.name)
		return err
	}
	stderr, err := ws.exec.StderrPipe()
	if err != nil {
		logging.Error("can not get stderr pipe for link [%s] name [%s]", ws.id, ws.name)
		return err
	}
	err = ws.exec.Start()
	if err != nil {
		logging.Error("can not start code-server for link [%s] name [%s]", ws.id, ws.name)
		return err
	}
	go ws.log(stdout, stderr)
	return nil
}

// Close close workspace
func (ws *Workspace) Close() {
	if ws.exec != nil && ws.exec.Process != nil {
		ws.exec.Process.Kill()
	}
	ws.remote.SendDisconnect(ws.target, ws.id)
}

func (ws *Workspace) log(stdout, stderr io.ReadCloser) {
	defer stdout.Close()
	defer stderr.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	watch := func(target io.Reader) {
		defer wg.Done()
		s := bufio.NewScanner(target)
		for s.Scan() {
			logging.Info("code-server [%s] [%s]: %s", ws.id, ws.name, s.Text())
		}
	}

	go watch(stdout)
	go watch(stderr)
	wg.Wait()
}

func (ws *Workspace) Forward() {

}
