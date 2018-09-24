package deduper

import (
	"container/list"
	"time"
)

type Keyer interface {
	// create a key for uniqueness check
	// should be unique, comparable, suitable for map
	Key() interface{}
}

type DeduperConfig struct {
	MaxEntries int
}

type Deduper struct {
	maxEntries int
	// map of Key()s for fast lookup
	lookup   map[interface{}]*list.Element
	deathrow *list.List
}

type item struct {
	value Keyer
	t     time.Time
}

func newItem(input Keyer) item {
	return item{input, time.Now()}
}

func New(c DeduperConfig) *Deduper {
	return &Deduper{
		maxEntries: c.MaxEntries,
		lookup:     make(map[interface{}]*list.Element),
		deathrow:   list.New()}
}

// Consume input and return:
//         original value in cache,
//         time of original value,
//         true if input is brand new, or false if it's already in cache
func (d *Deduper) Consume(input Keyer) (Keyer, time.Time, bool) {
	l := d.deathrow
	key := input.Key()
	if e, ok := d.lookup[key]; ok {
		// list is not empty and it's a match
		// promote that element
		if prev := e.Prev(); prev != nil {
			l.MoveBefore(e, prev)
		}
		it := e.Value.(item)
		return it.value, it.t, false
	}

	if l.Len() > d.maxEntries {
		// cleanup
		// remove back element as the rarest
		it := l.Remove(l.Back()).(item)
		delete(d.lookup, it.value.Key())
	}
	it := newItem(input)
	d.lookup[key] = l.PushFront(it)
	return it.value, it.t, true
}
