package eq

import (
	"container/list"
	"time"
)

// Config describes various settings of the ExpireQueue.
type Config struct {
	// TTL specifies the TTL of an item in the queue. While performing
	// Set operation the TTL will be checked for items in the bottom
	// of the queue. CleanN operation also relies on this setting. If
	// zero, no limit will be imposed on the queue.
	TTL time.Duration

	// MaxItems specifies max number of keys to store in the queue. If
	// there is more than MaxItems in a queue, Set operation will
	// expire the item in the bottom of the queue. If zero, no limit
	// will be imposed on the queue. CleanN also respects this
	// setting.
	MaxItems int

	// BackScan specifies maximum items will be scanned from bottom
	// to the top until non-expiring item will be met. All items found
	// are to be removed as expired on Set operation. Zero value is
	// the same as 1.
	BackScan int
}

// ExpireQueue implements basic Get/Set map operations in a form of a
// priority queue.
type ExpireQueue struct {
	max  int // MaxItems
	bs   int // back scan
	ttl  time.Duration
	elts map[interface{}]*list.Element
	row  *list.List
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// NewWithOpts return new instance of ExpireQueue with specified
// options.
func NewWithOpts(c *Config) *ExpireQueue {
	return &ExpireQueue{
		max:  c.MaxItems,
		ttl:  c.TTL,
		bs:   max(c.BackScan, 1),
		elts: make(map[interface{}]*list.Element),
		row:  list.New()}
}

// New is a shortcut for NewWithOpts with BackScan set to 1 and
// MaxItems set to 0.
func New(ttl time.Duration) *ExpireQueue {
	return NewWithOpts(&Config{TTL: ttl})
}

type box struct {
	updated time.Time
	k, v    interface{}
}

// SetTTL specifies new TTL for entries.
func (q *ExpireQueue) SetTTL(ttl time.Duration) {
	q.ttl = ttl
}

// Set sets key and value in a queue.
func (q *ExpireQueue) Set(k, v interface{}) {
	c := q.Back()
	now := time.Now()
	if c.IsNil() || (!q.IsFull() && !c.IsExpired(now)) {
		q.Push(k, v, now, nil)
	} else {
		q.Push(k, v, now, &c)
	}
}

// Get retrives value by key.
func (q *ExpireQueue) Get(k interface{}) (interface{}, bool) {
	c := q.NewCursor(k)
	if c.IsNil() {
		return nil, false
	}

	if c.IsExpired(time.Now()) {
		c.Delete()
		return nil, false
	}
	_, v := c.KeyValue()
	return v, true
}

// Delete removes key and value.
func (q *ExpireQueue) Delete(k interface{}) {
	if c := q.NewCursor(k); !c.IsNil() {
		c.Delete()
	}
}

// CleanN tries to pop up to n tail elements if they're expired.
func (q *ExpireQueue) CleanN(n int) {
	c := q.Back()
	now := time.Now()
	for i := 0; i < n; i++ {
		if c.IsNil() || !c.IsExpired(now) {
			return
		}
		c.Delete()
	}
}

// Count returns number of elements in the queue.
func (q *ExpireQueue) Count() int {
	return len(q.elts)
}
