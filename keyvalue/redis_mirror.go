package keyvalue

import (
	"context"
	"fmt"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/pubsub"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/util"
	"log"
	"runtime"
	"strings"
	"time"
)

const (
	// number of tuples to collect in bulk
	// for querying redis
	queryChannelBuf = 128

	// how often to send bulk redis queries
	queryDrainInterval = 100 * time.Millisecond
)

type RedisConfig struct {
	Network, Addr string
	DbIndex       int
	ScanCount     int
}

type tuple struct {
	k, v      string
	withValue bool
}

type RedisMirror struct {
	pool    *pool.Pool
	store   Map
	index   int
	scanCnt int
	input   chan tuple
	keys    map[string]time.Time
}

func getEvent(channel string) string {
	return strings.SplitN(channel, ":", 2)[1]
}

func logIfErr(prefix string, err error) {
	if err != nil {
		log.Println(prefix + ": " + err.Error())
	}
}

func (m *RedisMirror) queryRedis(key string) (string, error) {
	resp := m.pool.Cmd("GET", key)
	return resp.Str()
}

func NewRedisMirror(m Map, c RedisConfig) *RedisMirror {
	redisPool, err := pool.New(c.Network, c.Addr, 10)
	logIfErr("error initializing pool:", err)
	return &RedisMirror{
		pool:    redisPool,
		store:   m,
		input:   make(chan tuple, c.ScanCount),
		keys:    make(map[string]time.Time),
		index:   c.DbIndex,
		scanCnt: c.ScanCount}
}

func (m *RedisMirror) processQueries(ctx context.Context) {
	buf := make([]interface{}, 0, queryChannelBuf)
	ticker := time.NewTicker(queryDrainInterval)
	defer ticker.Stop()

	// LMDB requires to know the lock state of goroutine
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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
					logIfErr("map.put", m.store.Put(keys[i].(string), value))
				} else if r.IsType(redis.Nil) {
					logIfErr("map.del", m.store.Del(keys[i].(string)))
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-m.input:
			if t.withValue {
				logIfErr("map.put", m.store.Put(t.k, t.v))
				break
			}
			if buf = append(buf, t.k); len(buf) < cap(buf) {
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

func (m *RedisMirror) Get(key string) (string, error) {
	if value, err := m.store.Get(key); err == nil {
		return value, nil
	} else if value, err := m.queryRedis(key); err != nil {
		return "", err
	} else {
		t := tuple{k: key, v: value, withValue: true}
		select {
		case m.input <- t:
		default:
		}
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

func (m *RedisMirror) Scan() {
	scanner := util.NewScanner(m.pool,
		util.ScanOpts{Command: "SCAN", Count: m.scanCnt})
	for scanner.HasNext() {
		m.input <- tuple{k: scanner.Next(), withValue: false}
	}
}

func (m *RedisMirror) RedisMirror(ctx context.Context) {
	m.input = make(chan tuple, m.scanCnt)
	defer close(m.input)

	newctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go m.processQueries(newctx)

	// bootstrap scan
	go m.Scan()
MAIN_LOOP:
	for {
		if isClosed(ctx) {
			return
		}
		cl, err := m.pool.Get()
		if err != nil {
			log.Println("error getting connection:", err.Error())
			continue
		}

		eventFilter := fmt.Sprintf("__keyevent@%d__:*", m.index)
		subcl := pubsub.NewSubClient(cl)
		if resp := subcl.PSubscribe(eventFilter); resp.Err != nil {
			log.Println("error subscribing to events:", resp.Err.Error())
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
				log.Println("error receiving event:", resp.Err.Error())
				cl.Close()
				goto MAIN_LOOP
			} else if resp.Type != pubsub.Message {
				continue
			}

			key, event := resp.Message, getEvent(resp.Channel)
			switch event {
			case "expire":
				fallthrough
			case "expired":
				fallthrough
			case "del":
				m.input <- tuple{k: key, withValue: false}
				//logIfErr("map.del", m.store.Del(key))
			}
		}
	}
}
