package pool

import (
	"context"
	"natpass/code/network"
	"strings"
	"sync"
	"time"

	"github.com/lwch/logging"
)

type Conn struct {
	sync.RWMutex
	ID          string
	parent      *Pool
	conn        *network.Conn
	read        map[string]chan *network.Msg // link id => channel
	unknownRead chan *network.Msg            // read message without link
	write       chan *network.Msg            // link id => channel
}

func newConn(parent *Pool, conn *network.Conn, id string) *Conn {
	ret := &Conn{
		ID:          id,
		parent:      parent,
		conn:        conn,
		read:        make(map[string]chan *network.Msg),
		unknownRead: make(chan *network.Msg),
		write:       make(chan *network.Msg),
	}
	logging.Info("new connection: %s", ret.ID)
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
	conn.parent.onClose(conn.ID)
	logging.Error("connection %s closed", conn.ID)
}

func (conn *Conn) loopRead(cancel context.CancelFunc) {
	defer conn.Close()
	defer cancel()
	for {
		msg, err := conn.conn.ReadMessage(conn.parent.cfg.ReadTimeout)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			logging.Error("read message: %v", err)
			return
		}
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
		}
		conn.RLock()
		ch := conn.read[linkID]
		conn.RUnlock()
		if ch == nil {
			ch = conn.unknownRead
		}
		select {
		case ch <- msg:
		case <-time.After(conn.parent.cfg.ReadTimeout):
			logging.Error("write read channel for link %s timeouted", linkID)
			if ch == conn.unknownRead {
				continue
			}
			close(ch)
		}
	}
}

func loopWrite(conn *network.Conn, msg *network.Msg, timeout time.Duration) error {
	return conn.WriteMessage(msg, timeout)
}

func (conn *Conn) loopWrite(cancel context.CancelFunc) {
	defer conn.Close()
	defer cancel()
	for {
		msg := <-conn.write
		if msg == nil {
			return
		}
		msg.From = conn.ID
		err := loopWrite(conn.conn, msg, conn.parent.cfg.WriteTimeout)
		if err != nil {
			logging.Error("write message error on %s: %v", conn.ID, err)
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
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			conn.SendKeepalive()
		}
	}
}
