package eq

import (
	"container/list"
	"time"
)

// Cursor points to specific element of the queue.
type Cursor struct {
	q   *ExpireQueue
	elt *list.Element
}

// NewCursor creates a cursor at the specified key. If corresponding
// element doesn't exist return Nil element.
func (q *ExpireQueue) NewCursor(k interface{}) Cursor {
	e, ok := q.elts[k]
	if !ok {
		e = nil
	}

	return Cursor{q, e}
}

// Front returns front element of the queue.
func (q *ExpireQueue) Front() Cursor {
	return Cursor{q, q.row.Front()}
}

// Back returns rear element of the queue.
func (q *ExpireQueue) Back() Cursor {
	return Cursor{q, q.row.Back()}
}

// KeyValue returns key-value pair under the cursor.
func (c *Cursor) KeyValue() (k, v interface{}) {
	b := c.elt.Value.(box)
	return b.k, b.v
}

// IsExpired tells if the cursor has expired key-value relative to
// given timestamp.
func (c *Cursor) IsExpired(now time.Time) bool {
	if q := c.q; q.ttl != 0 {
		// there's a time limit and we check it.
		b := c.elt.Value.(box)
		return now.After(b.updated.Add(q.ttl))
	}

	return false
}

// IsNil tells if the cursor is a Nil element, i.e. not related to
// actual key/value in the queue.
func (c *Cursor) IsNil() bool {
	return c.elt == nil
}

// Next moves cursor to the back of the queue by one step.
func (c *Cursor) Next() Cursor {
	return Cursor{c.q, c.elt.Next()}
}

// Prev moves cursor to the top of the queue by one step.
func (c *Cursor) Prev() Cursor {
	return Cursor{c.q, c.elt.Prev()}
}

// Delete removes current key/value and returns a cursor right after
// given one of Nil element if it doesn't exist.
func (c *Cursor) Delete() Cursor {
	q, e := c.q, c.elt
	next := e.Next()
	b := q.row.Remove(e).(box)
	delete(q.elts, b.k)
	return Cursor{q, next}
}

// Get jumps to the cursor of the given key.
func (c *Cursor) Get(k interface{}) Cursor {
	return c.q.NewCursor(k)
}

// Push inserts new element at the top of the queue with given
// key/value pair, timestamp and a cursor from this queue which may be
// substituted. The reusing of the cursor relieves pressure on GC.
func (q *ExpireQueue) Push(k, v interface{}, t time.Time, c *Cursor) {
	b := box{k: k, v: v, updated: t}
	if c == nil {
		q.elts[k] = q.row.PushFront(b)
	} else {
		e := c.elt
		old := e.Value.(box)
		e.Value = b
		delete(q.elts, old.k)
		q.elts[k] = e
		q.row.MoveToFront(e)
	}
}

// MoveToFront updates timestamp in cursor to specified value and
// moves it up front.
func (c *Cursor) MoveToFront(t time.Time) Cursor {
	e := c.elt
	b := e.Value.(box)
	b.updated = t
	e.Value = b
	c.q.row.MoveToFront(e)
	return *c
}

// IsFull tells if the queue if full, i.e. its length is at the max.
func (q *ExpireQueue) IsFull() bool {
	return q.max > 0 && len(q.elts) >= q.max
}

// Set sets new value for the key under the cursor.
// No timestamp change involved.
func (c *Cursor) Set(v interface{}) {
	e := c.elt
	b := e.Value.(box)
	b.v = v
	e.Value = b
}
