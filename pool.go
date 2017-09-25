package rebytes

import (
	"errors"
	"sync"
)

// Pool represents a goroutine-safe pool of []byte values of a configurable
// capacity.  Because it contains a sync.Mutex, it must not be copied.
type Pool struct {
	mutex    sync.Mutex
	slices   [][]byte
	bytesCap int
	poolCap  int
}

// NewPool constructs a new pool of []byte.  The first argument is the capacity
// for []byte values retrieved from the pool.  The second argument is the
// maximum number of unused []byte values that can be stored in the pool.
func NewPool(bytesCap, poolCap int) *Pool {
	p := Pool{bytesCap: bytesCap, poolCap: poolCap}
	p.slices = make([][]byte, 0, poolCap)
	return &p
}

// Get returns a []byte with zero length and the pool's configured
// []byte capacity.  The []byte value is removed from the pool if available
// or constructed on the spot otherwise.
func (p *Pool) Get() []byte {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	var b []byte

	// Pop and return cached slice if available
	if len(p.slices) > 0 {
		b, p.slices = p.slices[len(p.slices)-1], p.slices[0:len(p.slices)-1]
		return b
	}

	// Otherwise return new slice
	b = make([]byte, 0, p.bytesCap)
	return b
}

// Put places a []byte into the pool.  The pool takes ownership and the
// provided []byte MUST NOT be used again until returned from the pool with
// Get().  The []byte is resliced to zero length.  NOTE: data is not zeroed
// out, only the length.
//
// If the argument is nil or is
// a []byte with a capacity that isn't the same as the pool's configured []byte
// capacity, an error is returned and the argument is not added to the pool.
func (p *Pool) Put(b []byte) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Error if slice is nil
	if b == nil {
		return errors.New("nil value not allowed for rebytes.Pool.Put()")
	}

	// Error if slice is wrong capacity
	if cap(b) != p.bytesCap {
		return errors.New("wrong capacity byte slice for rebytes.Pool.Put()")
	}

	// Drop slice if pool is full
	if len(p.slices) == p.poolCap {
		return nil
	}

	// Otherwise, reset length and push slice onto stack
	b = b[0:0]
	p.slices = append(p.slices, b)
	return nil
}

// Size returns the number of []byte values in the pool.
func (p *Pool) Size() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return len(p.slices)
}
