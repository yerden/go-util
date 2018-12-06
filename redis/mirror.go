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

func NewMirror(c MirrorConfig) *Mirror {
	return &Mirror{
		r: NewRedis(RedisConfig{
			Addr:    c.Addr,
			Network: c.Network,
			DbIndex: c.DbIndex}),
		store: new(sync.Map)}
}

func (m *Mirror) SyncMap() *sync.Map {
	return m.store
}

func (m *Mirror) kvHandler() func(k, v interface{}) {
	return func(k, v interface{}) {
		if v == nil {
			m.store.Delete(k)
		} else {
			m.store.Store(k, v)
		}
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
	return m.r.ConsumeScan(context.Background(), m.kvHandler())
}

func (m *Mirror) Mirror(ctx context.Context) error {
	return m.r.ConsumeEvents(ctx, m.kvHandler())
}
