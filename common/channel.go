package common

import (
	"errors"
)

type Channel chan interface{}

func NewChannel(n int) Channel {
	return make(chan interface{}, n)
}

var (
	ErrChannelFull = errors.New("no space in channel")
)

func (ch Channel) Enqueue(v interface{}) error {
	select {
	case ch <- v:
		return nil
	default:
		return ErrChannelFull
	}
}

func (ch Channel) Dequeue() (interface{}, bool) {
	select {
	case v := <-ch:
		return v, true
	default:
		return nil, false
	}
}
