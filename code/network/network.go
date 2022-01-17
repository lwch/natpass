package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
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
	sizeRead  [6]byte
}

// NewConn create connection
func NewConn(c net.Conn) *Conn {
	return &Conn{c: c}
}

// Close close connection
func (c *Conn) Close() {
	c.c.Close()
}

// ReadMessage read message with timeout
func (c *Conn) ReadMessage(timeout time.Duration) (*Msg, uint16, error) {
	c.lockRead.Lock()
	defer c.lockRead.Unlock()
	c.c.SetReadDeadline(time.Now().Add(timeout))
	_, err := io.ReadFull(c.c, c.sizeRead[:])
	if err != nil {
		return nil, 0, err
	}
	size := binary.BigEndian.Uint16(c.sizeRead[:])
	enc := binary.BigEndian.Uint32(c.sizeRead[2:])
	buf := make([]byte, size)
	_, err = io.ReadFull(c.c, buf)
	if err != nil {
		return nil, 0, err
	}
	if crc32.ChecksumIEEE(buf) != enc {
		return nil, 0, errChecksum
	}
	var msg Msg
	err = proto.Unmarshal(buf, &msg)
	if err != nil {
		return nil, 0, err
	}
	return &msg, size, nil
}

// WriteMessage write message with timeout
func (c *Conn) WriteMessage(m *Msg, timeout time.Duration) error {
	data, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	if len(data) > math.MaxUint16 {
		return errTooLong
	}
	buf := make([]byte, len(data)+len(c.sizeRead))
	binary.BigEndian.PutUint16(buf, uint16(len(data)))
	binary.BigEndian.PutUint32(buf[2:], crc32.ChecksumIEEE(data))
	copy(buf[len(c.sizeRead):], data)
	c.lockWrite.Lock()
	defer c.lockWrite.Unlock()
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
