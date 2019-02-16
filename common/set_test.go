package common

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"golang.org/x/sys/unix"
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

func ExampleNewSetInt() {
	b := NewSetInt(2, 1, 3)

	fmt.Println(b.Count() == 3 &&
		b.IsSet(1) &&
		b.IsSet(2) &&
		b.IsSet(3))
	// Output: true
}

func TestSet(t *testing.T) {
	assert := newAssert(t, false)

	b := new(SetInt)
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
	b := NewSetInt(0, 2, 4)

	hex, err := MarshalHex(b)
	assert(err == nil)
	assert(bytes.Equal(hex, []byte("15")))

	err = UnmarshalHex(b, []byte("15"))
	assert(err == nil)
	assert(b.Count() == 3)
	assert(b.IsSet(0))
	assert(b.IsSet(2))
	assert(b.IsSet(4))

	err = UnmarshalHex(b, []byte("j2"))
	assert(err != nil)

	b.Zero()
	assert(nil == UnmarshalHex(b, []byte("0ff00ff")))
	assert(b.Count() == 16)
	for i := 0; i < 8; i++ {
		assert(b.IsSet(i))
	}
	for i := 16; i < 23; i++ {
		assert(b.IsSet(i))
	}
}

func ExampleUnmarshalHex() {
	var b SetInt
	s := "1f" // 0, 1, 2, 3, 4

	err := UnmarshalHex(&b, []byte(s))
	if err != nil {
		return
	}

	var list []int
	SetIterate(&b, func(c int) {
		list = append(list, c)
	})

	fmt.Println(list)
	// Output: [0 1 2 3 4]
}

func ExampleMarshalHex() {
	var b SetInt

	b.Set(0)
	b.Set(1)
	b.Set(2)
	b.Set(3)
	b.Set(4)

	s, err := MarshalHex(&b)
	if err != nil {
		return
	}

	fmt.Println(string(s))
	// Output: 1f
}

func ExampleSet_Merge() {
	a := NewSetInt(0, 1)
	b := NewSetInt(1, 2)

	a.Merge(b)

	var list []int
	a.Iterate(func(c int) {
		list = append(list, c)
	})

	fmt.Println(list)
	// Output: [0 1 2]
}

func ExampleSet_Cut() {
	a := NewSetInt(0, 1)
	b := NewSetInt(1, 2)

	a.Cut(b)

	var list []int
	a.Iterate(func(c int) {
		list = append(list, c)
	})

	fmt.Println(list)
	// Output: [0]
}

func ExampleUnmarshalText() {
	var b SetInt
	if err := UnmarshalText(&b, []byte("1-4,6,7")); err != nil {
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

func ExampleMarshalText() {
	b := NewSetInt(0, 2, 3, 4, 6, 7, 11)

	text, err := MarshalText(b)
	if err != nil {
		return
	}
	fmt.Println(string(text))
	// Output: 0,2-4,6-7,11
}

func ExampleUnmarshalHex_UnixCPUSet() {
	var b unix.CPUSet

	if err := UnmarshalHex(&b, []byte("001f")); err != nil {
		return
	}

	fmt.Println(b.Count() == 5 &&
		b.IsSet(0) && b.IsSet(1) &&
		b.IsSet(2) && b.IsSet(3) &&
		b.IsSet(4))

	fmt.Println(b.IsSet(5))

	// Output:
	// true
	// false
}

func ExampleMarshalHex_UnixCPUSet() {
	var b unix.CPUSet

	b.Set(0)
	b.Set(2)
	b.Set(3)
	b.Set(4)

	txt, err := MarshalHex(&b)
	if err != nil {
		return
	}

	fmt.Println(string(txt))

	// Output:
	// 1d
}

func testSet(set Set) {
	x := rand.Int()
	set.Set(x)
	set.Clear(x)
}

func testLookup(set Set) {
	x := rand.Int()
	_ = set.IsSet(x)
}

func testIterate(set Set) {
	SetIterate(set, func(c int) {
		_ = c
	})
}

func BenchmarkSetInt(b *testing.B) {
	set := NewSetInt(0, 2, 4, 5, 6, 7, 8, 12)
	for i := 0; i < b.N; i++ {
		testSet(set)
	}
}

func BenchmarkSetMap(b *testing.B) {
	set := NewSetMap(0, 2, 4, 5, 6, 7, 8, 12)
	for i := 0; i < b.N; i++ {
		testSet(set)
	}
}

func BenchmarkSetInt0(b *testing.B) {
	set := NewSetInt()
	for i := 0; i < b.N; i++ {
		testSet(set)
	}
}

func BenchmarkSetMap0(b *testing.B) {
	set := NewSetMap()
	for i := 0; i < b.N; i++ {
		testSet(set)
	}
}

func BenchmarkLookupInt(b *testing.B) {
	set := NewSetInt(0, 2, 4, 5, 6, 7, 8, 12)
	for i := 0; i < b.N; i++ {
		testLookup(set)
	}
}

func BenchmarkLookupMap(b *testing.B) {
	set := NewSetMap(0, 2, 4, 5, 6, 7, 8, 12)
	for i := 0; i < b.N; i++ {
		testLookup(set)
	}
}

func BenchmarkLookupInt0(b *testing.B) {
	set := NewSetInt()
	for i := 0; i < b.N; i++ {
		testLookup(set)
	}
}

func BenchmarkLookupMap0(b *testing.B) {
	set := NewSetMap()
	for i := 0; i < b.N; i++ {
		testLookup(set)
	}
}

func BenchmarkIterateInt(b *testing.B) {
	set := NewSetInt(0, 2, 4, 5, 6, 7, 8, 12)
	for i := 0; i < b.N; i++ {
		testIterate(set)
	}
}

func BenchmarkIterateMap(b *testing.B) {
	set := NewSetMap(0, 2, 4, 5, 6, 7, 8, 12)
	for i := 0; i < b.N; i++ {
		testIterate(set)
	}
}

func BenchmarkLookupIntLarge(b *testing.B) {
	var elts []int
	for i := 0; i < 1000; i++ {
		elts = append(elts, i)
	}
	set := NewSetInt(elts...)
	for i := 0; i < b.N; i++ {
		testLookup(set)
	}
}

func BenchmarkLookupMapLarge(b *testing.B) {
	var elts []int
	for i := 0; i < 1000; i++ {
		elts = append(elts, i)
	}
	set := NewSetMap(elts...)
	for i := 0; i < b.N; i++ {
		testLookup(set)
	}
}
