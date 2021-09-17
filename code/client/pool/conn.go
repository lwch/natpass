package pool

import (
	"context"
	"natpass/code/network"
	"natpass/code/utils"
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
)

type Conn struct {
	sync.RWMutex
	Idx         uint32
	parent      *Pool
	conn        *network.Conn
	read        map[string]chan *network.Msg // link id => channel
	unknownRead chan *network.Msg            // read message without link
	write       chan *network.Msg            // link id => channel
}

func newConn(parent *Pool, conn *network.Conn, idx uint32) *Conn {
	ret := &Conn{
		Idx:         idx,
		parent:      parent,
		conn:        conn,
		read:        make(map[string]chan *network.Msg),
		unknownRead: make(chan *network.Msg),
		write:       make(chan *network.Msg),
	}
	logging.Info("new connection: %s-%d", ret.parent.cfg.ID, ret.Idx)
	ctx, cancel := context.WithCancel(context.Background())
	go ret.loopRead(cancel)
	go ret.loopWrite(cancel)
	go ret.keepalive(ctx)
	return ret
}

func (conn *Conn) hasLink(id string) bool {
	conn.RLock()
	defer conn.RUnlock()
	_, ok := conn.read[id]
	return ok
}

// AddLink attach read message
func (conn *Conn) AddLink(id string) {
	conn.Lock()
	if _, ok := conn.read[id]; !ok {
		conn.read[id] = make(chan *network.Msg)
	}
	conn.Unlock()
}

// RemoveLink detach read message
func (conn *Conn) RemoveLink(id string) {
	conn.Lock()
	ch := conn.read[id]
	if ch != nil {
		close(ch)
	}
	delete(conn.read, id)
	conn.Unlock()
}

// Close close connection
func (conn *Conn) Close() {
	conn.conn.Close()
	conn.Lock()
	for id, ch := range conn.read {
		close(ch)
		delete(conn.read, id)
	}
	conn.Unlock()
	if conn.unknownRead != nil {
		close(conn.unknownRead)
		conn.unknownRead = nil
	}
	if conn.write != nil {
		close(conn.write)
		conn.write = nil
	}
	conn.parent.onClose(conn.Idx)
	logging.Error("connection %s-%d closed", conn.parent.cfg.ID, conn.Idx)
}

func (conn *Conn) loopRead(cancel context.CancelFunc) {
	defer utils.Recover("loopRead")
	defer conn.Close()
	defer cancel()
	var timeout int
	for {
		msg, err := conn.conn.ReadMessage(conn.parent.cfg.ReadTimeout)
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
		timeout = 0
		var linkID string
		switch msg.GetXType() {
		case network.Msg_connect_req:
			linkID = msg.GetCreq().GetId()
		case network.Msg_connect_rep:
			linkID = msg.GetCrep().GetId()
		case network.Msg_disconnect:
			linkID = msg.GetXDisconnect().GetId()
		case network.Msg_forward:
			linkID = msg.GetXData().GetLid()
		case network.Msg_keepalive:
			continue
		}
		conn.RLock()
		ch := conn.read[linkID]
		conn.RUnlock()
		if ch == nil {
			ch = conn.unknownRead
		}
		ch <- msg
	}
}

func (conn *Conn) loopWrite(cancel context.CancelFunc) {
	defer utils.Recover("loopWrite")
	defer conn.Close()
	defer cancel()
	for {
		msg := <-conn.write
		if msg == nil {
			return
		}
		msg.From = conn.parent.cfg.ID
		msg.FromIdx = conn.Idx
		err := conn.conn.WriteMessage(msg, conn.parent.cfg.WriteTimeout)
		if err != nil {
			logging.Error("write message error on %s-%d: %v",
				conn.parent.cfg.ID, conn.Idx, err)
			return
		}
	}
}

// ChanRead get read channel from link id
func (conn *Conn) ChanRead(id string) <-chan *network.Msg {
	conn.RLock()
	defer conn.RUnlock()
	return conn.read[id]
}

func (conn *Conn) ChanUnknown() <-chan *network.Msg {
	return conn.unknownRead
}

func (conn *Conn) keepalive(ctx context.Context) {
	defer utils.Recover("keepalive")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			conn.SendKeepalive()
		}
	}
}
