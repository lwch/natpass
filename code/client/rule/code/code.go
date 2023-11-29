package code

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/client/rule"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/runtime"
)

// Code code-server handler
type Code struct {
	sync.RWMutex
	Name         string
	cfg          *global.Rule
	workspace    map[string]*Workspace
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// New new code-server handler
func New(cfg *global.Rule, readTimeout, writeTimeout time.Duration) *Code {
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

// OnDisconnect on disconnect message
func (code *Code) OnDisconnect(id string) {
	code.RLock()
	workspace := code.workspace[id]
	code.RUnlock()
	if workspace != nil {
		workspace.Close(false)
	}
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
	mux.HandleFunc("/info", code.Info)
	mux.HandleFunc("/forward/", pf(code.Forward))
	mux.HandleFunc("/", pf(code.Render))
	if code.cfg.LocalPort == 0 {
		code.cfg.LocalPort = global.GeneratePort()
		logging.Info("generate port for %s: %d", code.Name, code.cfg.LocalPort)
	}
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", code.cfg.LocalAddr, code.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}

func (code *Code) remove(id string) {
	code.Lock()
	delete(code.workspace, id)
	code.Unlock()
}

func (code *Code) new(conn *conn.Conn) (string, error) {
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for code-server: %s, err=%v",
			code.Name, err)
		return "", err
	}
	link := code.NewLink(id, code.cfg.Target, nil, conn).(*Workspace)
	conn.SendConnectReq(id, code.cfg)
	ch := conn.ChanRead(id)
	var repMsg *network.Msg
	for {
		var msg *network.Msg
		select {
		case msg = <-ch:
		case <-time.After(time.Minute):
			logging.Error("create code-server %s by rule %s failed, timtout", link.id, link.parent.Name)
			return "", errWaitingTimeout
		}
		if msg.GetXType() != network.Msg_connect_rep {
			conn.Requeue(id, msg)
			time.Sleep(code.readTimeout / 10)
			continue
		}
		rep := msg.GetCrep()
		if !rep.GetOk() {
			logging.Error("create code-server %s by rule %s failed, err=%s",
				link.id, link.parent.Name, rep.GetMsg())
			return "", errors.New(rep.GetMsg())
		}
		repMsg = msg
		break
	}
	logging.Info("create link %s for code-server rule [%s] from %s to %s",
		link.GetID(), code.cfg.Name,
		repMsg.GetTo(), repMsg.GetFrom())
	go link.localRead()
	return id, nil
}
