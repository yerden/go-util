package common

import (
	"bufio"
)

type Vector interface {
	Len() int
	Clear()
}

type Queue interface {
	Enqueue(interface{}) error
	Dequeue() (interface{}, bool)
}

type Stack interface {
	Push(interface{}) error
	Pop() (interface{}, bool)
}

type Scanner interface {
	Scan() bool
	Text() string
	Bytes() []byte
	Err() error
}

type ScanCloser interface {
	Scanner
	Close()
}

var _ Scanner = (*bufio.Scanner)(nil)
