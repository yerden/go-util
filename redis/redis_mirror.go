package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/pubsub"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/util"
	"strings"
	"sync"
	"time"
)

const (
	queryChannelBuf    = 128
	queryDrainInterval = 100 * time.Millisecond
)

var (
	ErrNotFound = errors.New("Key not found")
)

type MirrorConfig struct {
	Network, Addr string
	FormatKey     Unmarshaller
	FormatValue   Unmarshaller
	DbIndex       int
	ScanCount     int
}

type Mirror struct {
	pool     *pool.Pool
	store    *sync.Map
	fmtKey   Unmarshaller
	fmtValue Unmarshaller
	index    int
	scanCnt  int
	input    chan string
}

func getEvent(channel string) string {
	return strings.SplitN(channel, ":", 2)[1]
}

func (m *Mirror) queryRedis(key interface{}) (interface{}, error) {
	skey := key.(fmt.Stringer).String()

	if resp := m.pool.Cmd("GET", skey); resp.Err != nil {
		return nil, resp.Err
	} else if sval, err := resp.Str(); err != nil {
		return nil, err
	} else {
		return m.fmtValue.Unmarshal(sval), nil
	}
}

func NewMirror(c MirrorConfig) *Mirror {
	redisPool, err := pool.New(c.Network, c.Addr, 10)
	if err != nil {
		fmt.Println("error initializing pool:", err.Error())
	}
	store := new(sync.Map)
	return &Mirror{
		pool:     redisPool,
		store:    store,
		fmtKey:   c.FormatKey,
		fmtValue: c.FormatValue,
		index:    c.DbIndex,
		scanCnt:  c.ScanCount}
}

func (m *Mirror) processQueries(ctx context.Context, ch <-chan string) {
	buf := make([]interface{}, 0, queryChannelBuf)
	ticker := time.NewTicker(queryDrainInterval)
	defer ticker.Stop()

	getSet := func(keys []interface{}) {
		if len(keys) == 0 {
			return
		} else if resp := m.pool.Cmd("MGET", keys...); resp.Err != nil {
			return
		} else if array, err := resp.Array(); err != nil || len(array) != len(keys) {
			return
		} else {
			for i, r := range array {
				if value, err := r.Str(); err == nil {
					m.store.Store(
						m.fmtKey.Unmarshal(keys[i].(string)),
						m.fmtValue.Unmarshal(value))
				} else if r.IsType(redis.Nil) {
					m.store.Delete(
						m.fmtKey.Unmarshal(keys[i].(string)))
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case key := <-ch:
			if buf = append(buf, key); len(buf) < cap(buf) {
				break
			}
			getSet(buf)
			buf = buf[:0]
		case <-ticker.C:
			getSet(buf)
			buf = buf[:0]
		}
	}
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

func (m *Mirror) Scan() {
	scanner := util.NewScanner(m.pool,
		util.ScanOpts{Command: "SCAN", Count: m.scanCnt})
	for scanner.HasNext() {
		m.input <- scanner.Next()
	}
}

func (m *Mirror) Mirror(ctx context.Context) {
	m.input = make(chan string, m.scanCnt)
	defer close(m.input)

	newctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go m.processQueries(newctx, m.input)

	// bootstrap scan
	go m.Scan()
MAIN_LOOP:
	for {
		if isClosed(ctx) {
			return
		}
		cl, err := m.pool.Get()
		if err != nil {
			fmt.Println("error getting connection:", err.Error())
			continue
		}

		eventFilter := fmt.Sprintf("__keyevent@%d__:*", m.index)
		subcl := pubsub.NewSubClient(cl)
		if resp := subcl.PSubscribe(eventFilter); resp.Err != nil {
			fmt.Println("error subscribing to events:", err.Error())
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
				fmt.Println("error receiving event:", resp.Err.Error())
				cl.Close()
				goto MAIN_LOOP
			} else if resp.Type != pubsub.Message {
				continue
			}

			key, event := resp.Message, getEvent(resp.Channel)
			switch event {
			case "expire":
				m.input <- key
			case "expired":
				fallthrough
			case "del":
				m.store.Delete(m.fmtKey.Unmarshal(key))
			}
		}
	}
}
