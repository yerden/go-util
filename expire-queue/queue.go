package eq

import (
	"container/list"
	"time"
)

type ExpireQueue struct {
	ttl  time.Duration
	elts map[interface{}]*list.Element
	row  *list.List
	ts   *TimeSource
}

func New(ttl, res time.Duration) *ExpireQueue {
	return &ExpireQueue{
		ttl:  ttl,
		elts: make(map[interface{}]*list.Element),
		row:  list.New(),
		ts:   NewTimeSource(res)}
}

type box struct {
	updated int64
	k, v    interface{}
}

// return true if given element is expired and should be expunged.
func (q *ExpireQueue) isExpired(nano int64, e *list.Element) bool {
	b := e.Value.(box)
	now := time.Unix(0, nano)
	future := time.Unix(0, b.updated).Add(q.ttl)
	return now.After(future)
}

// return back element and true if the element is expired and should
// be expunged.
func (q *ExpireQueue) isPop(nano int64) (*list.Element, bool) {
	e := q.row.Back()
	ok := e != nil
	if ok {
		ok = q.isExpired(nano, e)
	}
	return e, ok
}

// Scans n elements starting from the back of the row, checks if
// they're expired.  Then returns true if at least one element is
// expired. All found expired elements removed from the 'elts' map.
// All found expired elements are also removed from the row except
// last expired one which is also returned.
func (q *ExpireQueue) popN(nano int64, n int) (*list.Element, bool) {
	var e, elt *list.Element
	var ok bool
	for i := 0; i < n; i++ {
		if elt, ok = q.isPop(nano); !ok {
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
	nano := q.ts.UnixNano()
	b := box{updated: nano, k: k, v: v}

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
	if e, ok := q.popN(nano, 1); ok {
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
	nano := q.ts.UnixNano()
	e, ok := q.elts[k]
	if !ok {
		return nil, false
	}

	if q.isExpired(nano, e) {
		b := q.row.Remove(e).(box)
		delete(q.elts, b.k)
		return nil, false
	}

	return e.Value.(box).v, true
}

func (q *ExpireQueue) CleanN(n int) {
	if e, ok := q.popN(q.ts.UnixNano(), n); ok {
		q.row.Remove(e)
	}
}

func (q *ExpireQueue) Count() int {
	return len(q.elts)
}
