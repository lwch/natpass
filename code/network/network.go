package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

var errTooLong = errors.New("too long")
var errChecksum = errors.New("invalid checksum")

// Conn network connection
type Conn struct {
	c         net.Conn
	lockRead  sync.Mutex
	lockWrite sync.Mutex
	sizeRead  [4]byte
}

// NewConn create connection
func NewConn(c net.Conn) *Conn {
	return &Conn{c: c}
}

// Close close connection
func (c *Conn) Close() {
	c.c.Close()
}

func sum(data []byte) uint16 {
	tmp := make([]uint16, len(data)/2)
	binary.Read(bytes.NewReader(data), binary.BigEndian, tmp)
	var ret uint16
	for _, v := range tmp {
		ret += v
	}
	return ret
}

// ReadMessage read message with timeout
func (c *Conn) ReadMessage(timeout time.Duration) (*Msg, error) {
	c.lockRead.Lock()
	defer c.lockRead.Unlock()
	c.c.SetReadDeadline(time.Now().Add(timeout))
	_, err := io.ReadFull(c.c, c.sizeRead[:])
	if err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint16(c.sizeRead[:])
	enc := binary.BigEndian.Uint16(c.sizeRead[2:])
	buf := make([]byte, size)
	_, err = io.ReadFull(c.c, buf)
	if err != nil {
		return nil, err
	}
	if sum(buf) != enc {
		return nil, errChecksum
	}
	var msg Msg
	err = proto.Unmarshal(buf, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// WriteMessage write message with timeout
func (c *Conn) WriteMessage(m *Msg, timeout time.Duration) error {
	c.lockWrite.Lock()
	defer c.lockWrite.Unlock()
	data, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	if len(data) > math.MaxUint16 {
		return errTooLong
	}
	buf := make([]byte, len(data)+len(c.sizeRead))
	binary.BigEndian.PutUint16(buf, uint16(len(data)))
	binary.BigEndian.PutUint16(buf[2:], sum(data))
	copy(buf[len(c.sizeRead):], data)
	c.c.SetWriteDeadline(time.Now().Add(timeout))
	_, err = io.Copy(c.c, bytes.NewReader(buf))
	return err
}

// RemoteAddr get connection remote address
func (c *Conn) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

// LocalAddr get connection local address
func (c *Conn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}
