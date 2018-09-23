package deduper

import (
	"container/list"
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

func New(c DeduperConfig) *Deduper {
	return &Deduper{
		maxEntries: c.MaxEntries,
		lookup:     make(map[interface{}]*list.Element),
		deathrow:   list.New()}
}

func (d *Deduper) Consume(input Keyer) Keyer {
	l := d.deathrow
	key := input.Key()
	if e, ok := d.lookup[key]; ok {
		// list is not empty and it's a match
		// promote that element
		if prev := e.Prev(); prev != nil {
			l.MoveBefore(e, prev)
		}
		return nil
	}

	if l.Len() > d.maxEntries {
		// cleanup
		// remove back element as the rarest
		v := l.Remove(l.Back())
		delete(d.lookup, v.(Keyer).Key())
	}
	d.lookup[key] = l.PushFront(input)
	return input
}
