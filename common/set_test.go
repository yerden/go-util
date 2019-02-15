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

func TestSet(t *testing.T) {
	assert := newAssert(t, false)

	b := new(Set)
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
	b.Set(104)
	b.Set(50)
	assert(b.Count() == 5)
	assert(b.IsSet(1))
	assert(b.IsSet(2))
	assert(b.IsSet(4))
	assert(b.IsSet(104))
	assert(b.IsSet(50))
}

func TestSetMarshal(t *testing.T) {
	assert := newAssert(t, false)
	var b Set

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

	b.Zero()
	assert(nil == b.UnmarshalHex([]byte("0ff00ff")))
	assert(b.Count() == 16)
	for i := 0; i < 8; i++ {
		assert(b.IsSet(i))
	}
	for i := 16; i < 23; i++ {
		assert(b.IsSet(i))
	}
}
