package network

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/natpass/code/network/encoding"
	"github.com/lwch/natpass/code/network/encoding/proto"
)

var errTooLong = errors.New("transport: too long")
var errChecksum = errors.New("transport: invalid checksum")
var errTimeout = errors.New("transport: timeout")

// Conn network connection
type Conn struct {
	c          net.Conn
	lockRead   sync.Mutex
	chWrite    chan []byte
	codec      encoding.Codec
	compressor encoding.Compressor
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewConn create connection
func NewConn(c net.Conn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	conn := &Conn{
		c:       c,
		chWrite: make(chan []byte, 1024),
		codec:   proto.New(),
		ctx:     ctx,
		cancel:  cancel,
	}
	go conn.loopWrite()
	return conn
}

// SetCompressor set compressor
func (c *Conn) SetCompressor(cp encoding.Compressor) *Conn {
	c.compressor = cp
	return c
}

// SetCodec set codec
func (c *Conn) SetCodec(cc encoding.Codec) *Conn {
	c.codec = cc
	return c
}

// Close close connection
func (c *Conn) Close() {
	c.c.Close()
	c.cancel()
}

type header struct {
	Size     uint16
	Checksum uint32
}

func (c *Conn) read(timeout time.Duration) ([]byte, error) {
	c.lockRead.Lock()
	defer c.lockRead.Unlock()
	c.c.SetReadDeadline(time.Now().Add(timeout))
	var hdr header
	err := binary.Read(c.c, binary.BigEndian, &hdr)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, hdr.Size)
	_, err = io.ReadFull(c.c, buf)
	if err != nil {
		return nil, err
	}
	if crc32.ChecksumIEEE(buf) != hdr.Checksum {
		return nil, errChecksum
	}
	return buf, nil
}

func (c *Conn) unserialize(data []byte) (*Msg, error) {
	if c.compressor != nil {
		dec, err := c.compressor.Decompress(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		var buffer bytes.Buffer
		_, err = io.Copy(&buffer, dec)
		if err != nil {
			return nil, err
		}
		data = buffer.Bytes()
	}
	var msg Msg
	err := c.codec.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// ReadMessage read message with timeout
func (c *Conn) ReadMessage(timeout time.Duration) (*Msg, uint16, error) {
	buf, err := c.read(timeout)
	if err != nil {
		return nil, 0, err
	}
	msg, err := c.unserialize(buf)
	if err != nil {
		return nil, 0, err
	}
	return msg, uint16(len(buf)), nil
}

func (c *Conn) serialize(msg *Msg) ([]byte, error) {
	data, err := c.codec.Marshal(msg)
	if err != nil {
		return nil, err
	}
	if c.compressor != nil {
		var buffer bytes.Buffer
		enc, err := c.compressor.Compress(&buffer)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(enc, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return buffer.Bytes(), nil
	}
	return data, nil
}

func (c *Conn) write(data []byte, timeout time.Duration) error {
	hdr := header{
		Size:     uint16(len(data)),
		Checksum: crc32.ChecksumIEEE(data),
	}
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(&buffer, bytes.NewReader(data))
	if err != nil {
		return err
	}
	select {
	case c.chWrite <- buffer.Bytes():
		return nil
	case <-time.After(timeout):
		return errTimeout
	}
}

// WriteMessage write message with timeout
func (c *Conn) WriteMessage(msg *Msg, timeout time.Duration) error {
	data, err := c.serialize(msg)
	if err != nil {
		return err
	}
	if len(data) > math.MaxUint16 {
		return errTooLong
	}
	return c.write(data, timeout)
}

// RemoteAddr get connection remote address
func (c *Conn) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

// LocalAddr get connection local address
func (c *Conn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

func (c *Conn) loopWrite() {
	defer c.Close()
	for {
		select {
		case <-c.ctx.Done():
			return
		case data := <-c.chWrite:
			_, err := io.Copy(c.c, bytes.NewReader(data))
			if err != nil {
				logging.Error("write data: %v", err)
				return
			}
		}
	}
}
