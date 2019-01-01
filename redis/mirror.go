package redis

import (
	"context"
	"errors"
	//"fmt"
	"sync"
)

var (
	ErrNotFound = errors.New("Key not found")
)

type MirrorConfig struct {
	Network, Addr string
	DbIndex       int
}

type Mirror struct {
	r     *Redis
	store *sync.Map
}

func NewMirror(c MirrorConfig) (*Mirror, error) {
	r, err := NewRedis(RedisConfig{
		Addr:    c.Addr,
		Network: c.Network,
		DbIndex: c.DbIndex})
	return &Mirror{r: r, store: new(sync.Map)}, err
}

func (m *Mirror) SyncMap() *sync.Map {
	return m.store
}

func (m *Mirror) kvHandler() TupleOp {
	return func(k, v interface{}) bool {
		if v == nil {
			m.store.Delete(k)
		} else {
			m.store.Store(k, v)
		}
		return true
	}
}

func (m *Mirror) Get(key interface{}) (interface{}, error) {
	if value, ok := m.store.Load(key); ok {
		return value, nil
	}
	value, err := m.r.Get(key.(string))
	if err != nil {
		return nil, err
	}
	m.store.Store(key, value)
	return value, nil
}

func (m *Mirror) Scan() error {
	s := m.r.NewScanner(200)
	return m.r.ConsumeScanner(context.Background(), s, m.kvHandler())
}

func (m *Mirror) Mirror(ctx context.Context) error {
	s := m.r.NewKeyEventSource()
	defer s.Close()
	return m.r.ConsumeScanner(ctx, s, m.kvHandler())
}
