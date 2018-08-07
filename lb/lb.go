package lb

import (
	//"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

type Selector int

const (
	RoundRobin Selector = iota
	Random
	TimeUnixNano
)

type Config struct {
	// Number of worker goroutines
	Workers int
	// default goroutine selector
	Mode Selector
	// Goroutine buffer size, i.e. pending jobs max number
	BufferSize int
}

type LoadBalancer struct {
	farm []chan func()
	stop chan bool
	// for round-robin selector
	counter uint64
	choose  func() int
}

func New(config Config) *LoadBalancer {
	lb := &LoadBalancer{
		stop: make(chan bool, config.Workers),
		farm: make([]chan func(), config.Workers)}

	for i, _ := range lb.farm {
		ch := make(chan func(), config.BufferSize)
		lb.farm[i] = ch
		go func(ch chan func()) {
			for f := range ch {
				f()
			}
			lb.stop <- true
		}(ch)
	}

	switch config.Mode {
	default:
		fallthrough
	case RoundRobin:
		lb.choose = func() int {
			u := atomic.AddUint64(&lb.counter, 1)
			return int(u % uint64(len(lb.farm)))
		}
	case Random:
		lb.choose = func() int {
			return rand.Int() % len(lb.farm)
		}
	case TimeUnixNano:
		lb.choose = func() int {
			return int(time.Now().UnixNano() % int64(len(lb.farm)))
		}
	}
	return lb
}

// Stop cluster's goroutines.
func (lb *LoadBalancer) Stop() {
	for _, ch := range lb.farm {
		close(ch)
	}
	// wait for goroutines exit
	for _, _ = range lb.farm {
		<-lb.stop
	}
}

// Perform a job on specified `Item`.
func (lb *LoadBalancer) Do(f func()) {
	if N := len(lb.farm); N > 0 {
		if i := lb.choose(); i >= 0 && i < N {
			lb.farm[i] <- f
		}
	}
}

// Perform a job on specified `Item` with specific goroutine.
func (lb *LoadBalancer) DoWith(i int, f func()) {
	if N := len(lb.farm); N > 0 {
		lb.farm[i%N] <- f
	}
}

func (lb *LoadBalancer) Workers() int {
	return len(lb.farm)
}
