package code

import (
	"fmt"
	"net/http"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/conn"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/runtime"
)

// Code code-server handler
type Code struct {
	Name string
	cfg  global.Rule
}

// New new code-server handler
func New(cfg global.Rule) *Code {
	return &Code{
		Name: cfg.Name,
		cfg:  cfg,
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
	mux.HandleFunc("/", pf(code.Forward))
	svr := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", code.cfg.LocalAddr, code.cfg.LocalPort),
		Handler: mux,
	}
	runtime.Assert(svr.ListenAndServe())
}
