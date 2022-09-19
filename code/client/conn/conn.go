package conn

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/client/global"
	"github.com/lwch/natpass/code/network"
	"github.com/lwch/natpass/code/utils"
	"github.com/lwch/runtime"
)

const dropBlockTimeout = 10 * time.Minute

// Conn connection
type Conn struct {
	sync.RWMutex
	cfg          *global.Configure
	conn         *network.Conn
	read         map[string]chan *network.Msg // link id => channel
	unknownRead  chan *network.Msg            // read message without link
	onDisconnect chan string
	write        chan *network.Msg
	lockDrop     sync.RWMutex
	drop         map[string]time.Time
	// runtime
	ctx    context.Context
	cancel context.CancelFunc
}

// New new connection
func New(cfg *global.Configure) *Conn {
	conn := &Conn{
		cfg:          cfg,
		read:         make(map[string]chan *network.Msg),
		unknownRead:  make(chan *network.Msg, 1024),
		onDisconnect: make(chan string, 1024),
		write:        make(chan *network.Msg, 10*1024*1024),
		drop:         make(map[string]time.Time),
	}
	runtime.Assert(conn.connect())
	conn.ctx, conn.cancel = context.WithCancel(context.Background())
	go conn.loopRead()
	go conn.loopWrite()
	go conn.keepalive()
	go conn.checkDrop()
	return conn
}

func (conn *Conn) connect() error {
	var dial net.Conn
	var err error
	if conn.cfg.UseSSL {
		if conn.cfg.SSLInsecure {
			rawConn, err := net.Dial("tcp", conn.cfg.Server)
			if err != nil {
				logging.Error("raw dial: %v", err)
				return err
			}
			cfg := new(tls.Config)
			cfg.InsecureSkipVerify = true
			dial = tls.Client(rawConn, cfg)
			err = dial.(*tls.Conn).Handshake()
			if err != nil {
				rawConn.Close()
			}
		} else {
			dial, err = tls.Dial("tcp", conn.cfg.Server, nil)
		}
	} else {
		dial, err = net.Dial("tcp", conn.cfg.Server)
	}
	if err != nil {
		logging.Error("dial: %v", err)
		return err
	}
	cn := network.NewConn(dial)
	err = writeHandshake(cn, conn.cfg)
	if err != nil {
		logging.Error("write handshake: %v", err)
		return err
	}
	logging.Info("%s connected", conn.cfg.Server)
	conn.conn = cn
	return nil
}

func (conn *Conn) close() {
	if conn.conn != nil {
		conn.conn.Close()
	}
}

func writeHandshake(conn *network.Conn, cfg *global.Configure) error {
	var msg network.Msg
	msg.XType = network.Msg_handshake
	msg.From = cfg.ID
	msg.To = "server"
	msg.Payload = &network.Msg_Hsp{
		Hsp: &network.HandshakePayload{
			Enc: cfg.Hasher.Hash(),
		},
	}
	return conn.WriteMessage(&msg, 5*time.Second)
}

func (conn *Conn) isDrop(linkID string) bool {
	conn.lockDrop.RLock()
	defer conn.lockDrop.RUnlock()
	_, ok := conn.drop[linkID]
	return ok
}

func (conn *Conn) getChan(linkID string) chan *network.Msg {
	conn.RLock()
	ch := conn.read[linkID]
	conn.RUnlock()
	if ch == nil {
		ch = conn.unknownRead
	}
	return ch
}

func (conn *Conn) hookDispatch(ch chan *network.Msg, msg *network.Msg) bool {
	switch msg.GetXType() {
	case network.Msg_disconnect:
		conn.lockDrop.Lock()
		conn.drop[msg.GetLinkId()] = time.Now().Add(dropBlockTimeout)
		conn.lockDrop.Unlock()
		conn.onDisconnect <- msg.GetLinkId()
		logging.Info("connection %s disconnected", msg.GetLinkId())
		return false
	}
	return true
}

func (conn *Conn) loopRead() {
	defer utils.Recover("loopRead")
	defer conn.close()
	defer conn.cancel()
	var timeout int
	run := func(msg *network.Msg) bool {
		timeout = 0
		if msg.GetXType() == network.Msg_keepalive {
			return true
		}
		logging.Debug("read message %s(%s) from %s",
			msg.GetXType().String(), msg.GetLinkId(), msg.GetFrom())
		linkID := msg.GetLinkId()
		if conn.isDrop(linkID) {
			return true
		}
		ch := conn.getChan(linkID)
		if !conn.hookDispatch(ch, msg) {
			return true
		}
		select {
		case ch <- msg:
		case <-time.After(conn.cfg.WriteTimeout):
			logging.Error("drop message: %s", msg.GetXType().String())
			conn.lockDrop.Lock()
			conn.drop[msg.GetLinkId()] = time.Now().Add(dropBlockTimeout)
			conn.lockDrop.Unlock()
		case <-conn.ctx.Done():
			return false
		}
		return true
	}
	for {
		msg, _, err := conn.conn.ReadMessage(conn.cfg.ReadTimeout)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				timeout++
				if timeout >= 60 {
					logging.Error("too many timeout times")
					return
				}
				continue
			}
			logging.Error("read message: %v", err)
			return
		}
		if !run(msg) {
			return
		}
	}
}

func (conn *Conn) loopWrite() {
	defer utils.Recover("loopWrite")
	defer conn.close()
	defer conn.cancel()
	for {
		var msg *network.Msg
		select {
		case msg = <-conn.write:
		case <-conn.ctx.Done():
			return
		}
		msg.From = conn.cfg.ID
		err := conn.conn.WriteMessage(msg, conn.cfg.WriteTimeout)
		if err != nil {
			logging.Error("write message error on %s: %v",
				conn.cfg.ID, err)
			continue
		}
	}
}

func (conn *Conn) keepalive() {
	defer utils.Recover("keepalive")
	defer conn.close()
	defer conn.cancel()
	tk := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-tk.C:
			conn.SendKeepalive()
		case <-conn.ctx.Done():
			return
		}
	}
}

// AddLink attach read message
func (conn *Conn) AddLink(id string) {
	logging.Info("add link %s", id)
	conn.Lock()
	if _, ok := conn.read[id]; !ok {
		conn.read[id] = make(chan *network.Msg, 1024)
	}
	conn.Unlock()
}

// Reset reset message next read
func (conn *Conn) Reset(id string, msg *network.Msg) {
	conn.RLock()
	ch := conn.read[id]
	conn.RUnlock()
	ch <- msg
}

// ChanRead get read channel from link id
func (conn *Conn) ChanRead(id string) <-chan *network.Msg {
	conn.RLock()
	defer conn.RUnlock()
	return conn.read[id]
}

// ChanUnknown get channel of unknown link id
func (conn *Conn) ChanUnknown() <-chan *network.Msg {
	return conn.unknownRead
}

// ChanDisconnect get channel of disconnect
func (conn *Conn) ChanDisconnect() <-chan string {
	return conn.onDisconnect
}

func (conn *Conn) checkDrop() {
	for {
		time.Sleep(time.Second)

		drops := make([]string, 0, len(conn.drop))
		conn.lockDrop.RLock()
		for k, t := range conn.drop {
			if time.Now().After(t) {
				drops = append(drops, k)
			}
		}
		conn.lockDrop.RUnlock()

		conn.lockDrop.Lock()
		for _, id := range drops {
			delete(conn.drop, id)
		}
		conn.lockDrop.Unlock()
	}
}

// Wait wait for connection closed
func (conn *Conn) Wait() {
	<-conn.ctx.Done()
}

// ChanClose close read chan
func (conn *Conn) ChanClose(id string) {
	conn.Lock()
	ch := conn.read[id]
	if ch != nil {
		close(ch)
	}
	delete(conn.read, id)
	conn.Unlock()

	conn.lockDrop.Lock()
	conn.drop[id] = time.Now().Add(dropBlockTimeout)
	conn.lockDrop.Unlock()
}
