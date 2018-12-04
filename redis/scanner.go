package redis

import (
	"github.com/mediocregopher/radix.v2/util"
)

type ChanScanner struct {
	ch  chan string
	fin chan bool
	key string
}

var _ util.Scanner = (*ChanScanner)(nil)

func NewChanScanner() *ChanScanner {
	cs := &ChanScanner{
		ch:  make(chan string, 32),
		fin: make(chan bool)}
	return cs
}

func (cs *ChanScanner) HasNext() bool {
	select {
	case cs.key = <-cs.ch:
		return true
	case <-cs.fin:
		return false
	}
}

func (cs *ChanScanner) Next() string {
	return cs.key
}

func (cs *ChanScanner) Err() error {
	return nil
}

func (cs *ChanScanner) Put(k string) {
	cs.ch <- k
}

func (cs *ChanScanner) Close() {
	close(cs.fin)
}
