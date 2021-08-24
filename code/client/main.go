package main

import (
	"flag"
	"fmt"
	"natpass/code/client/global"
	"natpass/code/client/pool"
	"natpass/code/client/tunnel"
	"natpass/code/network"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/lwch/daemon"
	"github.com/lwch/logging"
)

var (
	_VERSION       string = "0.0.0"
	_GIT_HASH      string
	_GIT_REVERSION string
	_BUILD_TIME    string
)

func showVersion() {
	fmt.Printf("version: v%s\ntime: %s\ncommit: %s.%s\n",
		_VERSION,
		_BUILD_TIME,
		_GIT_HASH, _GIT_REVERSION)
	os.Exit(0)
}

func main() {
	bak := flag.Bool("d", false, "backend running")
	pid := flag.String("pid", "", "pid file dir")
	user := flag.String("u", "", "daemon user")
	conf := flag.String("conf", "", "configure file path")
	version := flag.Bool("v", false, "show version info")
	flag.Parse()

	if *version {
		showVersion()
		os.Exit(0)
	}

	if len(*conf) == 0 {
		fmt.Println("missing -conf param")
		os.Exit(1)
	}

	if *bak {
		daemon.Start(0, *pid, *user, "-conf", *conf)
		return
	}

	cfg := global.LoadConf(*conf)

	logging.SetSizeRotate(cfg.LogDir, "np-cli", int(cfg.LogSize.Bytes()), cfg.LogRotate, true)
	defer logging.Flush()

	pl := pool.New(cfg)

	for _, t := range cfg.Tunnels {
		tn := tunnel.New(t)
		go tn.Handle(pl)
	}

	for i := 0; i < cfg.Links-pl.Size(); i++ {
		go func() {
			for {
				conn := pl.Get()
				if conn == nil {
					time.Sleep(time.Second)
					continue
				}
				for {
					msg := <-conn.ChanUnknown()
					if msg == nil {
						break
					}
					var linkID string
					switch msg.GetXType() {
					case network.Msg_connect_req:
						connect(pl, conn, msg.GetFrom(), msg.GetTo(), msg.GetCreq())
					case network.Msg_connect_rep:
						linkID = msg.GetCrep().GetId()
					case network.Msg_disconnect:
						linkID = msg.GetXDisconnect().GetId()
					case network.Msg_forward:
						linkID = msg.GetXData().GetLid()
					}
					if len(linkID) > 0 {
						logging.Error("link of %s not found, type=%s", linkID, msg.GetXType().String())
						continue
					}
				}
				logging.Info("connection %s exited", conn.ID)
				time.Sleep(time.Second)
			}
		}()
	}

	select {}
}

func connect(pool *pool.Pool, conn *pool.Conn, from, to string, req *network.ConnectRequest) {
	dial := "tcp"
	if req.GetXType() == network.ConnectRequest_udp {
		dial = "udp"
	}
	link, err := net.Dial(dial, fmt.Sprintf("%s:%d", req.GetAddr(), req.GetPort()))
	if err != nil {
		conn.SendConnectError(from, req.GetId(), err.Error())
		return
	}
	host, pt, _ := net.SplitHostPort(link.LocalAddr().String())
	port, _ := strconv.ParseUint(pt, 10, 16)
	tn := tunnel.New(global.Tunnel{
		Name:       req.GetName(),
		Target:     from,
		Type:       dial,
		LocalAddr:  host,
		LocalPort:  uint16(port),
		RemoteAddr: req.GetAddr(),
		RemotePort: uint16(req.GetPort()),
	})
	lk := tunnel.NewLink(tn, req.GetId(), from, link, conn)
	conn.SendConnectOK(from, req.GetId())
	lk.Forward()
	lk.OnWork <- struct{}{}
}
