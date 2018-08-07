package lb

import (
	"testing"
)

func TestLB1(t *testing.T) {
	lb := New(Config{
		Workers:    5,
		Mode:       RoundRobin,
		BufferSize: 32})
	defer lb.Stop()

	if lb.Workers() != 5 {
		t.FailNow()
	}

	ch := make(chan int)

	lb.Do(func() {
		ch <- 1
	})

	if <-ch != 1 {
		t.FailNow()
	}

	lb.DoWith(2, func() {
		ch <- 2
	})

	if <-ch != 2 {
		t.FailNow()
	}
}
