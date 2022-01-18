package shell

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/jkstack/natpass/code/client/global"
	"github.com/jkstack/natpass/code/client/pool"
	"github.com/jkstack/natpass/code/client/rule"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

// Shell shell handler
type Shell struct {
	sync.RWMutex
	Name  string
	cfg   global.Rule
	links map[string]*Link
}

// New new shell
func New(cfg global.Rule) *Shell {
	return &Shell{
		Name:  cfg.Name,
		cfg:   cfg,
		links: make(map[string]*Link),
	}
}

// NewLink new link
func (shell *Shell) NewLink(id, remote string, remoteIdx uint32, localConn net.Conn, remoteConn *pool.Conn) rule.Link {
	remoteConn.AddLink(id)
	link := &Link{
		parent:    shell,
		id:        id,
		target:    remote,
		targetIdx: remoteIdx,
		remote:    remoteConn,
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

// Handle handle shell
func (shell *Shell) Handle(pl *pool.Pool) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close shell: %s, err=%v", shell.Name, err)
		}
	}()
	pf := func(cb func(*pool.Pool, http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cb(pl, w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", pf(shell.New))
	mux.HandleFunc("/ws/", pf(shell.WS))
	mux.HandleFunc("/resize", pf(shell.Resize))
	mux.HandleFunc("/", shell.Render)
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
