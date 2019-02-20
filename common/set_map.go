package common

import (
	"sort"
)

// SetMap is a set of integer numbers. Basic set operations
// are possible such as SetInt, Clear, Zero, IsSet. Possible
// applications may be set of CPU cores, integer numbers
// storage etc.
//
// Implemented on top of map.
type SetMap struct {
	hash map[int]bool
}

// NewSetMap creates new set. elts is an array
// of integers. SetInt takes control over elts after
// this call so it may not be used thereafter.
func NewSetMap(elts ...int) *SetMap {
	b := &SetMap{make(map[int]bool)}
	for _, c := range elts {
		b.hash[c] = true
	}
	return b
}

// SetMap removes n to the set.
func (b *SetMap) Clear(c int) {
	delete(b.hash, c)
}

// Set adds n to the set.
func (b *SetMap) Set(c int) {
	b.hash[c] = true
}

// Count returns number of elements in set.
func (b *SetMap) Count() int {
	return len(b.hash)
}

// IsSet tells if n is in set.
func (b *SetMap) IsSet(c int) bool {
	_, ok := b.hash[c]
	return ok
}

// Zero clears out the set.
func (b *SetMap) Zero() {
	b.hash = make(map[int]bool)
}

// Iterate scrolls through members of set.
func (b *SetMap) Iterate(fn func(int)) {
	elts := make([]int, 0, b.Count())
	for c, _ := range b.hash {
		elts = append(elts, c)
	}
	sort.Ints(elts)
	for _, c := range elts {
		fn(c)
	}
}
