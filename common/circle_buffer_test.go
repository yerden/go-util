package common

import (
	//"fmt"
	"github.com/yerden/go-util/assert"
	"math/rand"
	"testing"
	"time"
)

func TestCBuffer(t *testing.T) {
	a := assert.New(t)
	buf := new(CBuffer)

	for i := 0; i < 100; i++ {
		buf.Enqueue(i)
	}

	a.Equal(buf.Len(), 100)
	a.True(buf.Len() <= buf.Cap())
	for i := 0; i < 100; i++ {
		v, ok := buf.Dequeue()
		a.True(ok)
		a.Equal(buf.Len(), 100-i-1)
		a.Equal(v.(int), i)
	}
	a.Equal(buf.Len(), 0)

	_, ok := buf.Dequeue()
	a.NotTrue(ok)

	// check that buffer doesn't grow on repeat
	oldCap := buf.Cap()
	for i := 0; i < 100; i++ {
		buf.Enqueue(i)
	}

	a.Equal(buf.Len(), 100)
	a.True(buf.Len() <= buf.Cap())
	for i := 0; i < 100; i++ {
		v, ok := buf.Dequeue()
		a.True(ok)
		a.Equal(buf.Len(), 100-i-1)
		a.Equal(v.(int), i)
	}
	a.Equal(buf.Len(), 0)
	a.Equal(buf.Cap(), oldCap)
}

func TestCBufferRW(t *testing.T) {
	a := assert.New(t)
	buf := &CBuffer{}
	rand.Seed(time.Now().Unix())

	buf.Enqueue(1)
	oldCap := buf.Cap()
	buf.Dequeue()

	for i := 0; i < 100; i++ {
		x := rand.Int()
		buf.Enqueue(x)
		v, ok := buf.Dequeue()
		a.True(ok)
		a.Equal(oldCap, buf.Cap())
		a.Equal(v.(int), x)
	}
}

func TestCBufferCornerCase1(t *testing.T) {
	a := assert.New(t)
	buf := &CBuffer{}
	res := []int{}

	i := 0

	// fill the buffer
	for buf.Len() < 16 {
		buf.Enqueue(i)
		i++
	}

	// read buffer partially to move cursors
	for k := 0; k < 10; k++ {
		v, ok := buf.Dequeue()
		a.True(ok)
		res = append(res, v.(int))
	}

	// fill the buffer again
	for buf.Len() < buf.Cap() {
		buf.Enqueue(i)
		i++
	}

	// ... and one value to overflow
	buf.Enqueue(i)

	// check values
	for {
		v, ok := buf.Dequeue()
		if !ok {
			break
		}
		res = append(res, v.(int))
	}

	for i, x := range res {
		a.Equal(i, x)
	}

	var k int
	for k, _ = range res {
		a.Equal(k, res[k])
	}
	a.Equal(k, i)
}

func TestCBufferCornerCase2(t *testing.T) {
	a := assert.New(t)
	buf := &CBuffer{}
	res := []int{}

	i := 0

	// fill the buffer
	for buf.Len() < 32 {
		buf.Enqueue(i)
		i++
	}

	// read buffer partially to move cursors
	for k := 0; k < 10; k++ {
		v, ok := buf.Dequeue()
		a.True(ok)
		res = append(res, v.(int))
	}

	// fill the buffer again
	for buf.Len() < buf.Cap() {
		buf.Enqueue(i)
		i++
	}

	// ... and one value to overflow
	buf.Enqueue(i)

	// check values
	for {
		v, ok := buf.Dequeue()
		if !ok {
			break
		}
		res = append(res, v.(int))
	}

	var k int
	for k, _ = range res {
		a.Equal(k, res[k])
	}
	a.Equal(k, i)
}
