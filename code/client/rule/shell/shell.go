package shell

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/client/rule"
	"github.com/lwch/runtime"
)

// Shell shell handler
type Shell struct {
	sync.RWMutex
	Name         string
	cfg          *global.Rule
	links        map[string]*Link
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// New new shell
func New(cfg *global.Rule, readTimeout, writeTimeout time.Duration) *Shell {
	return &Shell{
		Name:         cfg.Name,
		cfg:          cfg,
		links:        make(map[string]*Link),
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

// NewLink new link
func (shell *Shell) NewLink(id, remote string, localConn net.Conn, remoteConn *conn.Conn) rule.Link {
	remoteConn.AddLink(id)
	link := &Link{
		parent: shell,
		id:     id,
		target: remote,
		remote: remoteConn,
	}
	shell.Lock()
	shell.links[link.id] = link
	shell.Unlock()
	return link
}

// GetName get shell rule name
func (shell *Shell) GetName() string {
	return shell.Name
}

// GetTypeName get shell rule type name
func (shell *Shell) GetTypeName() string {
	return "shell"
}

// GetTarget get target of this rule
func (shell *Shell) GetTarget() string {
	return shell.cfg.Target
}

// GetLinks get rule links
func (shell *Shell) GetLinks() []rule.Link {
	ret := make([]rule.Link, 0, len(shell.links))
	shell.RLock()
	for _, link := range shell.links {
		ret = append(ret, link)
	}
	shell.RUnlock()
	return ret
}

// GetRemote get remote target name
func (shell *Shell) GetRemote() string {
	return shell.cfg.Target
}

// GetPort get listen port
func (shell *Shell) GetPort() uint16 {
	return shell.cfg.LocalPort
}

// OnDisconnect on disconnect message
func (shell *Shell) OnDisconnect(id string) {
	shell.RLock()
	link := shell.links[id]
	shell.RUnlock()
	if link != nil {
		link.Close(false)
	}
}

// Handle handle shell
func (shell *Shell) Handle(c *conn.Conn) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close shell: %s, err=%v", shell.Name, err)
		}
	}()
	pf := func(cb func(*conn.Conn, http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cb(c, w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", pf(shell.New))
	mux.HandleFunc("/ws/", pf(shell.WS))
	mux.HandleFunc("/resize", pf(shell.Resize))
	mux.HandleFunc("/", shell.Render)
	if shell.cfg.LocalPort == 0 {
		shell.cfg.LocalPort = global.GeneratePort()
		logging.Info("generate port for %s: %d", shell.Name, shell.cfg.LocalPort)
	}
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", shell.cfg.LocalAddr, shell.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}

func (shell *Shell) remove(id string) {
	shell.Lock()
	delete(shell.links, id)
	shell.Unlock()
}
