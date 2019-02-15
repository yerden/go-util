package common

import (
	"bytes"
	"fmt"
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

func ExampleSet_UnmarshalHex() {
	var b Set
	s := "1f" // 0, 1, 2, 3, 4

	err := b.UnmarshalHex([]byte(s))
	if err != nil {
		return
	}

	var list []int
	b.Iterate(func(c int) {
		list = append(list, c)
	})

	fmt.Println(list)
	// Output: [0 1 2 3 4]
}

func ExampleSet_MarshalHex() {
	var b Set

	b.Set(0)
	b.Set(1)
	b.Set(2)
	b.Set(3)
	b.Set(4)

	s, err := b.MarshalHex()
	if err != nil {
		return
	}

	fmt.Println(string(s))
	// Output: 1f
}

func ExampleSet_Merge() {
	var a, b Set

	a.Set(0)
	a.Set(1)

	b.Set(1)
	b.Set(2)

	a.Merge(&b)

	var list []int
	a.Iterate(func(c int) {
		list = append(list, c)
	})

	fmt.Println(list)
	// Output: [0 1 2]
}

func ExampleSet_Cut() {
	var a, b Set

	a.Set(0)
	a.Set(1)

	b.Set(1)
	b.Set(2)

	a.Cut(&b)

	var list []int
	a.Iterate(func(c int) {
		list = append(list, c)
	})

	fmt.Println(list)
	// Output: [0]
}

func ExampleSet_UnmarshalText() {
	var b Set
	if err := b.UnmarshalText([]byte("1-4,6,7")); err != nil {
		return
	}

	fmt.Println(b.Count() == 6 &&
		b.IsSet(1) &&
		b.IsSet(2) &&
		b.IsSet(3) &&
		b.IsSet(4) &&
		b.IsSet(6) &&
		b.IsSet(7))
	// Output: true
}

func ExampleSet_MarshalText() {
	var b Set
	b.Set(0)
	b.Set(2)
	b.Set(3)
	b.Set(4)
	b.Set(6)
	b.Set(7)
	b.Set(11)

	text, err := b.MarshalText()
	if err != nil {
		return
	}
	fmt.Println(string(text))
	// Output: 0,2-4,6-7,11
}
