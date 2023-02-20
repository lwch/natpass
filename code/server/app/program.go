package app

import (
	"crypto/tls"
	"fmt"
	"net"
	rt "runtime"

	"github.com/common-nighthawk/go-figure"
	"github.com/kardianos/service"
	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/server/global"
	"github.com/lwch/natpass/code/server/handler"
	"github.com/lwch/runtime"
)

type program struct {
	cfg *global.Configure
}

func newProgram() *program {
	return &program{}
}

func (p *program) setConfigure(cfg *global.Configure) *program {
	p.cfg = cfg
	return p
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	// initialize logging
	stdout := true
	if rt.GOOS == "windows" {
		stdout = false
	}
	logging.SetSizeRotate(logging.SizeRotateConfig{
		Dir:         p.cfg.LogDir,
		Name:        "np-svr",
		Size:        int64(p.cfg.LogSize.Bytes()),
		Rotate:      p.cfg.LogRotate,
		WriteStdout: stdout,
		WriteFile:   true,
	})
	defer logging.Flush()

	fg := figure.NewFigure("NatPass", "alligator2", false)
	figure.Write(&logging.DefaultLogger, fg)
	logging.DefaultLogger.Write(nil)

	// go func() {
	// 	http.ListenAndServe(":7878", nil)
	// }()

	var l net.Listener
	if len(p.cfg.TLSCrt) > 0 && len(p.cfg.TLSKey) > 0 {
		cert, err := tls.LoadX509KeyPair(p.cfg.TLSCrt, p.cfg.TLSKey)
		runtime.Assert(err)
		l, err = tls.Listen("tcp", fmt.Sprintf(":%d", p.cfg.Listen), &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
		runtime.Assert(err)
		logging.Info("listen on %d", p.cfg.Listen)
	} else {
		var err error
		l, err = net.Listen("tcp", fmt.Sprintf(":%d", p.cfg.Listen))
		runtime.Assert(err)
	}

	p.serve(l)
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) serve(l net.Listener) {
	h := handler.New(p.cfg)
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go h.Handle(conn)
	}
}
