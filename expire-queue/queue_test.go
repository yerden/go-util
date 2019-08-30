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

func TestSet(t *testing.T) {
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

func BenchmarkSet(b *testing.B) {
	// assert := Assert(b, true)

	q := New(time.Microsecond)

	for i := 0; i < b.N; i++ {
		q.Set(i, i+1)
	}
}
