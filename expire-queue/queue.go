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

// return true if given element is expired and should be expunged.
func (q *ExpireQueue) isExpired(now time.Time, e *list.Element) bool {
	if q.max > 0 && e.Next() == nil && len(q.elts) > q.max {
		// we're at the bottom, and there's a limit on maximum number
		// of items and we hit the limit.
		return true
	}

	if q.ttl != 0 {
		// there's a time limit and we check it.
		b := e.Value.(box)
		return now.After(b.updated.Add(q.ttl))
	}

	return false
}

// return back element and true if the element is expired and should
// be expunged.
func (q *ExpireQueue) isPop(now time.Time) (*list.Element, bool) {
	e := q.row.Back()
	ok := e != nil
	if ok {
		ok = q.isExpired(now, e)
	}
	return e, ok
}

// Scans n elements starting from the back of the row, checks if
// they're expired.  Then returns true if at least one element is
// expired. All found expired elements removed from the 'elts' map.
// All found expired elements are also removed from the row except
// last expired one which is also returned.
func (q *ExpireQueue) popN(now time.Time, n int) (*list.Element, bool) {
	var e, elt *list.Element
	var ok bool
	for i := 0; i < n; i++ {
		if elt, ok = q.isPop(now); !ok {
			break
		}

		b := elt.Value.(box)
		delete(q.elts, b.k)
		if e != nil {
			q.row.Remove(e)
		}
		e = elt
	}

	return e, e != nil
}

// Flags to use in SmartSet which define the behaviour in case key is
// already in a queue. No flags mean no action.
const (
	// If key is already present, replace key's value with a new one.
	Replace uint = 1 << iota

	// If key is already present, update key's timestamp and
	// move it to the top of the queue. Value is not changed.
	Revive
)

// SmartSet sets new k/v. flags defines what to do if the key is
// present in the map.
func (q *ExpireQueue) SmartSet(k, v interface{}, flags uint) {
	// try find it
	if e, ok := q.elts[k]; ok {
		if flags == 0 {
			return
		}
		// update value in existing element
		b := e.Value.(box)
		{
			if (flags & Replace) != 0 {
				b.v = v
			}
			if (flags & Revive) != 0 {
				b.updated = time.Now()
				q.row.MoveToFront(e)
			}
		}
		e.Value = b

		// we could check for expiring values in the back but we
		// choose to do that on insertion because we'd like to relieve
		// pressure on GC. In this situation we don't require new
		// element so we spare possibly expiring element until a new
		// k/v is inserted.
		return
	}

	now := time.Now()
	b := box{k: k, v: v, updated: now}
	// maybe we can reuse some dead's clothes
	if e, ok := q.popN(now, q.bs); ok {
		// update value in existing element
		e.Value = b
		q.row.MoveToFront(e)

		// and set new element in the map
		q.elts[k] = e
		return
	}

	// add new element
	q.elts[k] = q.row.PushFront(b)
}

// Set sets key and value in a queue.
func (q *ExpireQueue) Set(k, v interface{}) {
	q.SmartSet(k, v, Replace|Revive)
}

// Get retrives value by key.
func (q *ExpireQueue) Get(k interface{}) (interface{}, bool) {
	now := time.Now()
	e, ok := q.elts[k]
	if !ok {
		return nil, false
	}

	if q.isExpired(now, e) {
		b := q.row.Remove(e).(box)
		delete(q.elts, b.k)
		return nil, false
	}

	return e.Value.(box).v, true
}

// Delete removes key and value.
func (q *ExpireQueue) Delete(k interface{}) {
	if e, ok := q.elts[k]; ok {
		delete(q.elts, k)
		q.row.Remove(e)
	}
}

// CleanN tries to pop up to n tail elements if they're expired.
func (q *ExpireQueue) CleanN(n int) {
	if e, ok := q.popN(time.Now(), n); ok {
		q.row.Remove(e)
	}
}

// Count returns number of elements in the queue.
func (q *ExpireQueue) Count() int {
	return len(q.elts)
}
