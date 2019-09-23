package eq

import (
	"testing"
	"time"
)

func TestCursor(t *testing.T) {
	assert := Assert(t, true)

	q := New(time.Second)

	q.Set("apple", "fruit")
	q.Set("tomato", "vegetable")
	q.Set("fennel", "plant")

	c := q.NewCursor("apple")
	_, v := c.KeyValue()
	assert(!c.IsNil())
	assert(v.(string) == "fruit")

	c = c.Get("tomato")
	_, v = c.KeyValue()
	assert(!c.IsNil())
	assert(v.(string) == "vegetable")

	c = q.Front()
	k, v := c.KeyValue()
	assert(!c.IsNil())
	assert(k.(string) == "fennel")
	assert(v.(string) == "plant")

	c = q.Back()
	k, v = c.KeyValue()
	assert(!c.IsNil())
	assert(k.(string) == "apple")
	assert(v.(string) == "fruit")

	c = c.MoveToFront(time.Now())
	c = q.Front()
	k, v = c.KeyValue()
	assert(!c.IsNil())
	assert(k.(string) == "apple")
	assert(v.(string) == "fruit")

	n := 0
	for c := q.Front(); !c.IsNil(); c = c.Next() {
		n++
	}
	assert(n == 3, n)

	c.Delete()
	c = q.NewCursor("apple")
	assert(c.IsNil())
}
