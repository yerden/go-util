package keyvalue

import (
	"sync"
)

type SyncMap sync.Map

var _ Map = (*SyncMap)(nil)

func (m *SyncMap) GetN(k, v []string) ([]string, error) {
	if v == nil {
		v = make([]string, len(k))
	}

	for i, key := range k {
		val, ok := (*sync.Map)(m).Load(key)
		if !ok {
			return v[:i], ErrNotFound
		}
		v[i] = val.(string)
	}

	return v, nil
}

func (m *SyncMap) PutN(k, v []string) error {
	for i, key := range k {
		(*sync.Map)(m).Store(key, v[i])
	}
	return nil
}

func (m *SyncMap) DelN(k []string) error {
	for _, key := range k {
		(*sync.Map)(m).Delete(key)
	}
	return nil
}
