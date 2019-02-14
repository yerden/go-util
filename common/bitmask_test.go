package common

import (
	"bytes"
	"testing"
)

func newAssert(t *testing.T, fail bool) func(bool) {
	return func(expected bool) {
		if t.Helper(); !expected {
			t.Error("Something's not right")
			if fail {
				t.FailNow()
			}
		}
	}
}

func TestBitmask(t *testing.T) {
	assert := newAssert(t, false)

	b := new(Bitmask)
	b.Set(0)
	assert(b.IsSet(0))
	assert(!b.IsSet(1))
	assert(!b.IsSet(2))
	b.Set(2)
	assert(b.IsSet(2))

	b.Clear(2)
	assert(!b.IsSet(2))

	b.Set(1)
	b.Set(3)
	b.Set(4)
	b.Set(7)
	b.Zero()
	assert(!b.IsSet(0))
	assert(b.Count() == 0)

	b.Set(2)
	b.Set(4)
	b.Set(1)
	assert(b.Count() == 3)
	assert(b.IsSet(1))
	assert(b.IsSet(2))
	assert(b.IsSet(4))
}

func TestBitmask2(t *testing.T) {
	assert := newAssert(t, false)
	var b Bitmask

	b.Set(0)
	b.Set(2)
	b.Set(4)

	hex, err := b.MarshalHex()
	assert(err == nil)
	assert(bytes.Equal(hex, []byte("15")))

	err = b.UnmarshalHex([]byte("15"))
	assert(err == nil)
	assert(b.Count() == 3)
	assert(b.IsSet(0))
	assert(b.IsSet(2))
	assert(b.IsSet(4))

	err = b.UnmarshalHex([]byte("j2"))
	assert(err != nil)
}
