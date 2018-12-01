package keyvalue

import (
	"errors"
	//"fmt"
	//"time"
)

var (
	ErrNotFound = errors.New("key not found")
)

// string -> string interface
type Map interface {
	// get a value of a given key
	Get(k string) (string, error)

	// put a key with value into map
	Put(k, v string) error

	// delete a key
	Del(k string) error

	// populate array with arbitrary
	// keys from map, ErrNotFound if no keys
	Sample([]string) (int, error)
}
