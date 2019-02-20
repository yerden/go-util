package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
)

const (
	byteBits = 8
)

type Set interface {
	Set(int)
	Clear(int)
	IsSet(int) bool
	Zero()
	Count() int
}

type SetIterable interface {
	Set

	// set iterator, it must return elements
	// in sorted order
	Iterate(func(int))
}

// SetIterate scrolls through Set via native Iterate method
// or using a fallback to try-and-check non-negative index
// loop.
func SetIterate(b Set, fn func(int)) {
	if bi, ok := b.(SetIterable); ok {
		bi.Iterate(fn)
		return
	}

	for i, cnt := 0, 0; cnt < b.Count(); i++ {
		if b.IsSet(i) {
			fn(i)
			cnt++
		}
	}
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

// MarshalHex marshals SetIterable's internal representation
// to hexadecimal big-endian string.
func MarshalHex(b Set) ([]byte, error) {
	var mask []byte
	SetIterate(b, func(c int) {
		mask = set(c, mask)
	})
	if len(mask) == 0 {
		return []byte{'0'}, nil
	}

	buf := bytes.NewBuffer(make([]byte, 0, 32))
	enc := hex.NewEncoder(buf)
	_, err := enc.Write(mask)
	return buf.Bytes(), err

}

// UnmarshalHex unmarshals hexadecimal big-endian string
// into Set's internal representation.
func UnmarshalHex(b Set, text []byte) (err error) {
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

// UnmarshalText unmarshals comma/hyphen separated list of elements
// into Set's internal representation.
func UnmarshalText(b Set, text []byte) error {
	b.Zero()
	var err error
	var token []byte
	for len(text) > 0 {
		if i := bytes.IndexByte(text, ','); i < 0 {
			token, text = text, text[:0]
		} else {
			token, text = text[:i], text[i+1:]
		}
		if len(token) == 0 {
			return fmt.Errorf("invalid format")
		}
		s, e := 0, 0
		if i := bytes.IndexByte(token, '-'); i < 0 {
			_, err = fmt.Sscanf(string(token), "%d", &s)
			e = s
		} else {
			_, err = fmt.Sscanf(string(token), "%d-%d", &s, &e)
		}

		if err != nil {
			return err
		}
		for i := s; i <= e; i++ {
			b.Set(i)
		}
	}

	return nil
}

// MarshalText marshals SetIterable internal representation
// into to comma/hyphen separated list of elements.
func MarshalText(b Set) ([]byte, error) {
	var buf bytes.Buffer
	token := make([]int, 0, 2)

	save := func() {
		if len(token) == 0 {
			return
		}
		if buf.Len() > 0 {
			fmt.Fprint(&buf, ",")
		}
		if len(token) == 1 {
			fmt.Fprintf(&buf, "%d", token[0])
		} else if len(token) == 2 {
			fmt.Fprintf(&buf, "%d-%d", token[0], token[1])
		}
	}

	SetIterate(b, func(c int) {
		switch len(token) {
		case 0:
			token = append(token, c)
			return
		case 1:
			if c == token[0]+1 {
				token = append(token, c)
				return
			}
		case 2:
			if c == token[1]+1 {
				token[1] = c
				return
			}
		}
		save()
		token = append(token[:0], c)
	})
	save()
	return buf.Bytes(), nil
}

// Merge add all members of src to set.
func Merge(dst, src Set) {
	SetIterate(src, dst.Set)
}

// Cut removes all members of src from set.
func Cut(dst, src Set) {
	SetIterate(src, dst.Clear)
}
