package code

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

// Code code-server handler
type Code struct {
	sync.RWMutex
	Name         string
	cfg          global.Rule
	workspace    map[string]*Workspace
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// New new code-server handler
func New(cfg global.Rule, readTimeout, writeTimeout time.Duration) *Code {
	return &Code{
		Name:         cfg.Name,
		cfg:          cfg,
		workspace:    make(map[string]*Workspace),
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

// GetName get code-server rule name
func (code *Code) GetName() string {
	return code.Name
}

// GetTypeName get code-server rule type name
func (code *Code) GetTypeName() string {
	return "code-server"
}

// GetPort get listen port
func (code *Code) GetPort() uint16 {
	return code.cfg.LocalPort
}

// GetTarget get target of this rule
func (code *Code) GetTarget() string {
	return code.cfg.Target
}

// GetLinks get rule links
func (code *Code) GetLinks() []rule.Link {
	ret := make([]rule.Link, 0, len(code.workspace))
	code.RLock()
	for _, link := range code.workspace {
		ret = append(ret, link)
	}
	code.RUnlock()
	return ret
}

// GetRemote get remote target name
func (code *Code) GetRemote() string {
	return code.cfg.Target
}

// NewLink new link
func (code *Code) NewLink(id, remote string, localConn net.Conn, remoteConn *conn.Conn) rule.Link {
	remoteConn.AddLink(id)
	ws := newWorkspace(code, id, code.cfg.Name, remote, remoteConn)
	code.Lock()
	code.workspace[ws.id] = ws
	code.Unlock()
	return ws
}

// Handle handle code-server
func (code *Code) Handle(c *conn.Conn) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("close code-server: %s, err=%v", code.Name, err)
		}
	}()
	pf := func(cb func(*conn.Conn, http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cb(c, w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", pf(code.New))
	mux.HandleFunc("/", pf(code.Forward))
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", code.cfg.LocalAddr, code.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}
