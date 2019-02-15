package common

import (
	"bytes"
	"encoding/hex"
	// "fmt"
	"io"
	"sort"
)

const (
	byteBits = 8
)

// Set is a set of integer numbers. Basic set operations
// are possible such as Set, Clear, Zero, IsSet. Possible
// applications may be set of CPU cores, integer numbers
// storage etc.
type Set struct {
	shift []int
}

func set(n int, mask []byte) []byte {
	id := n / byteBits
	rem := uint(n - id*byteBits)
	if id >= len(mask) {
		mask = append(make([]byte, id-len(mask)+1), mask...)
	}
	id = len(mask) - 1 - id
	mask[id] |= 1 << rem
	return mask
}

func get(mask []byte) []int {
	var res []int
	for i, _ := range mask {
		sym := mask[len(mask)-1-i]

		for k := 0; k < byteBits; k++ {
			if sym&(1<<uint(k)) != 0 {
				res = append(res, byteBits*i+k)
			}
		}
	}
	return res
}

func reverse(data []byte) {
	for s, e := 0, len(data)-1; s < e; s, e = s+1, e-1 {
		data[s], data[e] = data[e], data[s]
	}
}

// NewSet creates new set. elts is an array
// of integers. Set takes control over elts after
// this call so it may not be used thereafter.
func NewSet(elts []int) *Set {
	sort.Ints(elts)
	return &Set{elts}
}

func (b *Set) find(n int) (idx int, found bool) {
	idx = sort.SearchInts(b.shift, n)
	found = idx < len(b.shift) && b.shift[idx] == n
	return
}

// Set adds n to the set.
func (b *Set) Set(n int) {
	if i, found := b.find(n); !found {
		tail := b.shift[i:]
		b.shift = append(b.shift, n)
		if len(tail) > 0 {
			copy(b.shift[i+1:], tail)
			b.shift[i] = n
		}
	}
}

// Set removes n to the set.
func (b *Set) Clear(n int) {
	if i, found := b.find(n); found {
		b.shift = append(b.shift[:i], b.shift[i+1:]...)
	}
}

// IsSet tells if n is in set.
func (b *Set) IsSet(n int) bool {
	_, found := b.find(n)
	return found
}

// Zero clears out the set.
func (b *Set) Zero() {
	b.shift = b.shift[:0]
}

// Count returns number of elements in set.
func (b *Set) Count() int {
	return len(b.shift)
}

// Iterate scrolls through members of set.
func (b *Set) Iterate(fn func(int)) {
	for _, c := range b.shift {
		fn(c)
	}
}

// Merge add all members of src to set.
func (b *Set) Merge(src *Set) {
	src.Iterate(b.Set)
}

// Cut removes all members of src from set.
func (b *Set) Cut(src *Set) {
	src.Iterate(b.Clear)
}

// MarshalHex marshals set internal representation
// to hexadecimal big-endian string.
func (b *Set) MarshalHex() ([]byte, error) {
	var mask []byte
	for i, _ := range b.shift {
		mask = set(b.shift[len(b.shift)-i-1], mask)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 32))
	enc := hex.NewEncoder(buf)
	_, err := enc.Write(mask)
	return buf.Bytes(), err
}

// UnmarshalHex unmarshals hexadecimal big-endian string
// into its own internal representation.
func (b *Set) UnmarshalHex(text []byte) (err error) {
	// padding
	if len(text)&0x1 != 0 {
		text = append([]byte{'0'}, text...)
	}

	dec := hex.NewDecoder(bytes.NewReader(text))
	buf := bytes.NewBuffer(make([]byte, 0, 16))
	if _, err = io.Copy(buf, dec); err == nil {
		b.Zero()
		for _, n := range get(buf.Bytes()) {
			b.Set(n)
		}
	}

	return err
}
