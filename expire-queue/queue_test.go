package eq

import (
	"testing"
	"time"
)

func Assert(t testing.TB, fail bool) func(bool, ...interface{}) {
	return func(expected bool, v ...interface{}) {
		if !expected {
			t.Helper()
			if t.Error(v...); fail {
				t.FailNow()
			}
		}
	}
}

func TestSetTTL(t *testing.T) {
	assert := Assert(t, true)

	q := New(time.Second)
	assert(q != nil)

	q.Set(1, 2)
	q.Set(2, 3)

	v, ok := q.Get(1)
	assert(v != nil)
	assert(ok)
	assert(v.(int) == 2)

	v, ok = q.Get(2)
	assert(v != nil)
	assert(ok)
	assert(v.(int) == 3)

	v, ok = q.Get(3)
	assert(!ok)

	time.Sleep(time.Second)

	v, ok = q.Get(1)
	assert(!ok)
	v, ok = q.Get(2)
	assert(!ok)
}

func TestSetMaxItems(t *testing.T) {
	assert := Assert(t, true)

	max := 100
	c := &Config{MaxItems: max}
	q := NewWithOpts(c)
	assert(q != nil)

	for n := 0; n < max; n++ {
		q.Set(n, n)
	}

	for n := 0; n < max; n++ {
		_, ok := q.Get(n)
		assert(ok)
	}

	q.Set(max, max)
	_, ok := q.Get(0)
	assert(!ok)
}

func BenchmarkSet(b *testing.B) {
	// assert := Assert(b, true)

	q := New(time.Microsecond)

	for i := 0; i < b.N; i++ {
		q.Set(i, i+1)
	}
}
