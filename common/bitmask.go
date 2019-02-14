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

type Bitmask struct {
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

func (b *Bitmask) find(n int) (idx int, found bool) {
	idx = sort.SearchInts(b.shift, n)
	found = idx < len(b.shift) && b.shift[idx] == n
	return
}

func (b *Bitmask) Set(n int) {
	if i, found := b.find(n); !found {
		tail := b.shift[i:]
		b.shift = append(b.shift, n)
		if len(tail) > 0 {
			copy(b.shift[i+1:], tail)
			b.shift[i] = n
		}
	}
}

func (b *Bitmask) Clear(n int) {
	if i, found := b.find(n); found {
		b.shift = append(b.shift[:i], b.shift[i+1:]...)
	}
}

func (b *Bitmask) IsSet(n int) bool {
	_, found := b.find(n)
	return found
}

func (b *Bitmask) Zero() {
	b.shift = b.shift[:0]
}

func (b *Bitmask) Count() int {
	return len(b.shift)
}

func (b *Bitmask) Iterate(fn func(int)) {
	for _, c := range b.shift {
		fn(c)
	}
}

func (b *Bitmask) MarshalHex() ([]byte, error) {
	var mask []byte
	for _, n := range b.shift {
		mask = set(n, mask)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 32))
	enc := hex.NewEncoder(buf)
	_, err := enc.Write(mask)
	return buf.Bytes(), err
}

func (b *Bitmask) UnmarshalHex(text []byte) (err error) {
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
