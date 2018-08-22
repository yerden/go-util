package lb

import (
	"math/rand"
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

func TestLB2(t *testing.T) {
	var sum, sum1 int
	workers := 5
	ch := make(chan int, workers)
	// test data
	data := make([]int, 100)
	for i, _ := range data {
		data[i] = rand.Int()
	}

	// direct sum
	for _, i := range data {
		sum += i
	}

	// do it with LoadBalancer
	getF := func(pos, off int) func() {
		return func() {
			var s int
			for _, i := range data[pos : pos+off] {
				s += i
			}
			ch <- s
		}
	}

	lb := New(Config{
		Workers:    workers,
		Mode:       RoundRobin,
		BufferSize: 32})
	for i := 0; i < workers; i++ {
		lb.Do(getF(20*i, 20))
	}
	for i := 0; i < workers; i++ {
		sum1 += <-ch
	}

	if sum != sum1 {
		t.FailNow()
	}
}
