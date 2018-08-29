package lb

import (
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

const (
	RoundRobin int = iota
	Random
	TimeUnixNano
)

func NewSelector(seltype int) Selector {
	switch seltype {
	case RoundRobin:
		fallthrough
	default:
		return rr{new(uint64)}
	case Random:
		return random{}
	case TimeUnixNano:
		return timesel{}
	}
}
