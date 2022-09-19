package hash

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"hash"
	"math"
	"sync"
	"time"
)

// Hasher hasher for handshake
type Hasher struct {
	sync.Mutex
	period uint
	h      hash.Hash
}

// New create hasher
func New(secret string, period uint) *Hasher {
	if period == 0 {
		period = 30
	}
	return &Hasher{
		period: period,
		h:      hmac.New(sha512.New, []byte(secret)),
	}
}

// Hash hash func
func (h *Hasher) Hash() []byte {
	now := time.Now()
	i := math.Floor(float64(now.Unix()) / float64(h.period))
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	h.Lock()
	defer h.Unlock()
	h.h.Reset()
	h.h.Write(buf[:])
	ret := h.h.Sum(nil)
	return ret
}
