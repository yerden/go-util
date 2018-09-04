package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/pubsub"
	//"github.com/mediocregopher/radix.v2/redis"
	"strings"
	"sync"
)

var (
	ErrNotFound = errors.New("Key not found")
)

type KVFormatter interface {
	ToKey(string) interface{}
	ToValue(string) interface{}
	FromKey(interface{}) string
	FromValue(interface{}) string
}

type DefaultKVFormatter struct{}

func (x DefaultKVFormatter) ToKey(in string) interface{} {
	return in
}

func (x DefaultKVFormatter) ToValue(in string) interface{} {
	return in
}

func (x DefaultKVFormatter) FromKey(in interface{}) string {
	return in.(string)
}

func (x DefaultKVFormatter) FromValue(in interface{}) string {
	return in.(string)
}

type MirrorConfig struct {
	Network, Addr string
	Formatter     KVFormatter
	DbIndex       int
}

type Mirror struct {
	pool      *pool.Pool
	store     *sync.Map
	formatter KVFormatter
	index     int
}

func (m *Mirror) keyEvents(f string) string {
	return fmt.Sprintf("__keyevent@%d__:%s", m.index, f)
}

func getEvent(channel string) string {
	return strings.SplitN(channel, ":", 2)[1]
}

func (m *Mirror) queryRedis(key interface{}) (interface{}, error) {
	skey := m.formatter.FromKey(key)
	sval, err := m.queryRedisFmt(skey)
	if err != nil {
		return nil, err
	}
	return m.formatter.ToValue(sval), nil
}

func (m *Mirror) queryRedisFmt(key string) (string, error) {
	if resp := m.pool.Cmd("GET", key); resp.Err != nil {
		return "", resp.Err
	} else {
		return resp.Str()
	}
}

func NewMirror(c MirrorConfig) *Mirror {
	redisPool, err := pool.New(c.Network, c.Addr, 10)
	if err != nil {
		fmt.Println("error initializing pool:", err.Error())
	}
	store := new(sync.Map)
	return &Mirror{redisPool, store, c.Formatter, c.DbIndex}
}

func (m *Mirror) SyncMap() *sync.Map {
	return m.store
}

func (m *Mirror) Get(key interface{}) (interface{}, error) {
	if value, ok := m.store.Load(key); ok {
		return value, nil
	} else if value, err := m.queryRedis(key); err != nil {
		return nil, err
	} else {
		m.store.Store(key, value)
		return value, nil
	}
}

func isClosed(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (m *Mirror) ProcessEvents(ctx context.Context) {
MAIN_LOOP:
	for {
		if isClosed(ctx) {
			return
		}
		cl, err := m.pool.Get()
		if err != nil {
			continue
		}

		subcl := pubsub.NewSubClient(cl)
		if resp := subcl.PSubscribe(m.keyEvents("*")); resp.Err != nil {
			cl.Close()
			continue
		}

		for {
			if isClosed(ctx) {
				cl.Close()
				return
			}
			resp := subcl.Receive()
			if resp.Timeout() {
				continue
			} else if resp.Err != nil {
				cl.Close()
				goto MAIN_LOOP
			} else if resp.Type != pubsub.Message {
				continue
			}

			key := resp.Message
			switch event := getEvent(resp.Channel); event {
			case "expire":
				if value, err := m.queryRedisFmt(key); err == nil {
					m.store.Store(
						m.formatter.ToKey(key),
						m.formatter.ToValue(value))
				}
			case "expired":
				fallthrough
			case "del":
				m.store.Delete(m.formatter.ToKey(key))
			}
		}
	}
}
