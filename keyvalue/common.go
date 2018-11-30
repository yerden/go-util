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
	// k: supplied keys array
	// v: maybe used as a buffer for returned values
	//    if not supplied newly allocated buffer is returned.
	// returned values and error
	// if err == nil, all values were found
	// if err == ErrNotFound, returned slice contains values found up to error
	// if err != nil, internal error happened
	GetN(k, v []string) ([]string, error)

	// Put bulk key-values to map
	// k: supplied keys array
	// v: supplied values array, len(k) == len(v)
	PutN(k, v []string) error

	// Delete bulk key-values from map
	DelN(k []string) error
}

func MapGet(m Map, k string) (string, error) {
	v := []string{""}
	v, err := m.GetN([]string{k}, v[:0])
	return v[0], err
}

func MapPut(m Map, k, v string) error {
	return m.PutN([]string{k}, []string{v})
}

func MapDel(m Map, k string) error {
	return m.DelN([]string{k})
}

func errFromBool(ok bool) error {
	if ok {
		return nil
	}
	return ErrNotFound
}
