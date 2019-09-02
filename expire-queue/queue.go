package eq

import (
	"container/list"
	"time"
)

type ExpireQueue struct {
	ttl  time.Duration
	elts map[interface{}]*list.Element
	row  *list.List
}

func New(ttl time.Duration) *ExpireQueue {
	return &ExpireQueue{
		ttl:  ttl,
		elts: make(map[interface{}]*list.Element),
		row:  list.New()}
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
	b := e.Value.(box)
	return now.After(b.updated.Add(q.ttl))
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

func (q *ExpireQueue) Set(k, v interface{}) {
	now := time.Now()
	b := box{updated: now, k: k, v: v}

	// try find it
	if e, ok := q.elts[k]; ok {
		// update value in existing element
		e.Value = b
		q.row.MoveToFront(e)

		// we could check for expiring values in the back but we
		// choose to do that on insertion because we'd like to relieve
		// pressure on GC. In this situation we don't require new
		// element so we spare possibly expiring element until a new
		// k/v is inserted.
		return
	}

	// maybe we can reuse some dead's clothes
	if e, ok := q.popN(now, 1); ok {
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

func (q *ExpireQueue) CleanN(n int) {
	if e, ok := q.popN(time.Now(), n); ok {
		q.row.Remove(e)
	}
}

func (q *ExpireQueue) Count() int {
	return len(q.elts)
}
