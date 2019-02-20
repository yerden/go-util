package common

import (
	"sort"
)

// SetInt is a set of integer numbers. Basic set operations
// are possible such as SetInt, Clear, Zero, IsSet. Possible
// applications may be set of CPU cores, integer numbers
// storage etc.
//
// Implemented on top of sorted int array
type SetInt struct {
	shift []int
}

// NewSetInt creates new set. elts is an array
// of integers. SetInt takes control over elts after
// this call so it may not be used thereafter.
func NewSetInt(elts ...int) *SetInt {
	sort.Ints(elts)
	return &SetInt{elts}
}

func (b *SetInt) find(n int) (idx int, found bool) {
	idx = sort.SearchInts(b.shift, n)
	found = idx < len(b.shift) && b.shift[idx] == n
	return
}

// Set adds n to the set.
func (b *SetInt) Set(n int) {
	if i, found := b.find(n); !found {
		tail := b.shift[i:]
		b.shift = append(b.shift, n)
		if len(tail) > 0 {
			copy(b.shift[i+1:], tail)
			b.shift[i] = n
		}
	}
}

// SetInt removes n to the set.
func (b *SetInt) Clear(n int) {
	if i, found := b.find(n); found {
		b.shift = append(b.shift[:i], b.shift[i+1:]...)
	}
}

// IsSet tells if n is in set.
func (b *SetInt) IsSet(n int) bool {
	_, found := b.find(n)
	return found
}

// Zero clears out the set.
func (b *SetInt) Zero() {
	b.shift = b.shift[:0]
}

// Count returns number of elements in set.
func (b *SetInt) Count() int {
	return len(b.shift)
}

// Iterate scrolls through members of set.
func (b *SetInt) Iterate(fn func(int)) {
	for _, c := range b.shift {
		fn(c)
	}
}
