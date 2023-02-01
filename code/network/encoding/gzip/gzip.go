package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"sync"

	"github.com/lwch/natpass/code/network/encoding"
	"github.com/lwch/runtime"
)

type writer struct {
	*gzip.Writer
	pool *sync.Pool
}

// Close close write and put writer to pool
func (w *writer) Close() error {
	w.pool.Put(w)
	return w.Writer.Close()
}

type reader struct {
	*gzip.Reader
	pool *sync.Pool
}

// Close close reader and put reader to pool
func (r *reader) Close() error {
	r.pool.Put(r)
	return r.Reader.Close()
}

type compressor struct {
	level      int
	poolWriter [gzip.BestCompression]sync.Pool
	poolReader sync.Pool
}

// New create compressor
func New(level ...int) (encoding.Compressor, error) {
	if len(level) > 0 {
		if level[0] < 0 || level[0] > gzip.BestCompression {
			return nil, fmt.Errorf("invalid gzip compress level: %d", level[0])
		}
	} else {
		level = append(level, 6)
	}
	ret := new(compressor)
	ret.level = level[0]
	for i := 0; i < gzip.BestCompression; i++ {
		ret.poolWriter[i].New = func() interface{} {
			w, err := gzip.NewWriterLevel(io.Discard, i)
			runtime.Assert(err)
			return &writer{Writer: w, pool: &ret.poolWriter[i]}
		}
	}
	ret.poolReader.New = func() interface{} {
		r, err := gzip.NewReader(io.NopCloser(nil))
		runtime.Assert(err)
		return &reader{Reader: r, pool: &ret.poolReader}
	}
	return ret, nil
}

// Compress gzip compress
func (c *compressor) Compress(w io.Writer) (io.WriteCloser, error) {
	pw := c.poolWriter[c.level].Get().(*writer)
	pw.Writer.Reset(w)
	return pw, nil
}

// Decompress gzip decompress
func (c *compressor) Decompress(r io.Reader) (io.ReadCloser, error) {
	pr := c.poolReader.Get().(*reader)
	pr.Reader.Reset(r)
	return pr, nil
}

// SetLevel set compress level
func (c *compressor) SetLevel(level int) error {
	if level < 0 || level > gzip.BestCompression {
		return fmt.Errorf("invalid gzip compress level: %d", level)
	}
	c.level = level
	return nil
}
