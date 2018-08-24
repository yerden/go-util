package lb

import (
	//"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// select index out of range [0...max)
type Selector interface {
	Select(int) int
}

// round robin selector
type rr struct {
	cnt *uint64
}

func (s rr) Select(max int) int {
	u := atomic.AddUint64(s.cnt, 1)
	return int(u % uint64(max))
}

type random struct{}

func (s random) Select(max int) int {
	return rand.Int() % max
}

type timesel struct{}

func (s timesel) Select(max int) int {
	return int(time.Now().UnixNano() % int64(max))
}

var (
	RoundRobin   = Selector(rr{new(uint64)})
	Random       = Selector(random{})
	TimeUnixNano = Selector(timesel{})
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
	mode Selector
}

func New(config Config) *LoadBalancer {
	lb := &LoadBalancer{
		stop: make(chan bool, config.Workers),
		mode: config.Mode,
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
	lb.DoWith(lb.mode.Select(len(lb.farm)), f)
}

// Perform a job on specified `Item` with specific goroutine.
func (lb *LoadBalancer) DoWith(i int, f func()) {
	if N := len(lb.farm); N > 0 {
		lb.farm[i%N] <- f
	}
}

// Perform a bunch of jobs.
func (lb *LoadBalancer) DoBulk(ff []func()) {
	lb.DoBulkWith(lb.mode.Select(len(lb.farm)), ff)
}

// Perform a bunch of jobs with specific goroutine.
func (lb *LoadBalancer) DoBulkWith(i int, ff []func()) {
	lb.DoWith(i, func() {
		for _, x := range ff {
			x()
		}
	})
}

func (lb *LoadBalancer) Workers() int {
	return len(lb.farm)
}
