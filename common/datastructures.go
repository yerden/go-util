package common

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
