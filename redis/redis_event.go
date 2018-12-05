package redis

import (
	"context"
	"fmt"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/pubsub"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/util"
	"io"
	"log"
	"strings"
	"time"
)

type RedisConfig struct {
	Network, Addr string
	DbIndex       int
}

// event types
const (
	EventExpire = iota
	EventExpired
	EventDel
)

const (
	queryChannelBuf    = 128
	queryDrainInterval = 100 * time.Millisecond
	scanCount          = 200
)

type Redis struct {
	pool  *pool.Pool
	index int
}

type EventSource interface {
	HasNext() bool
	Err() error
	Next() string
	Type() int
	Close()
}

var _ EventSource = (*events)(nil)
var _ util.Scanner = EventSource(nil)

func getEvent(channel string) string {
	return strings.SplitN(channel, ":", 2)[1]
}

type events struct {
	r      *Redis
	filter string
	cl     *redis.Client
	subcl  *pubsub.SubClient

	// next event
	err error
	typ int
	key string
}

func (e *events) Type() int {
	return e.typ
}

func (e *events) Next() string {
	return e.key
}

func (e *events) Err() error {
	return e.err
}

func logIfErr(prefix string, err error) {
	if err != nil {
		log.Println(prefix + ": " + err.Error())
	}
}

func NewRedis(c RedisConfig) *Redis {
	// XXX: "If an error is encountered an empty
	// (but still usable) pool is returned alongside
	// that error"
	redisPool, _ := pool.New(c.Network, c.Addr, 10)
	return &Redis{pool: redisPool, index: c.DbIndex}
}

func (r *Redis) Get(key string) (string, error) {
	return r.pool.Cmd("GET", key).Str()
}

func (r *Redis) NewKeyEventSource() EventSource {
	e := &events{
		r:      r,
		filter: fmt.Sprintf("__keyevent@%d__:*", r.index)}
	return e
}

func (e *events) disconnect() {
	if e.cl != nil {
		e.cl.Close()
		e.cl = nil
	}
	e.subcl = nil
}

func (e *events) Close() {
	e.disconnect()
}

func (e *events) connect() bool {
	e.cl, e.err = e.r.pool.Get()
	if e.err != nil {
		return false
	}

	e.subcl = pubsub.NewSubClient(e.cl)
	resp := e.subcl.PSubscribe(e.filter)
	e.err = resp.Err
	return e.err == nil
}

func (e *events) HasNext() bool {
	for {
		if e.subcl == nil && !e.connect() {
			return false
		}
		resp := e.subcl.Receive()
		if resp.Timeout() {
			// "You can use the Timeout() method on
			// SubResp to easily determine if that
			// is the case. If this is the case you
			// can call Receive again to continue
			// listening for publishes."
			continue
		} else if resp.Err == io.EOF {
			// XXX: sometimes redis connection closes with EOF,
			// fetch new one and retry
			e.disconnect()
			if !e.connect() {
				return false
			}
			continue
		} else if e.err = resp.Err; e.err != nil {
			return false
		} else if resp.Type != pubsub.Message {
			continue
		}

		key, event := resp.Message, getEvent(resp.Channel)
		e.err = nil
		e.key = key
		switch event {
		case "expire":
			e.typ = EventExpire
		case "expired":
			e.typ = EventExpired
		case "del":
			e.typ = EventDel
		}
		return true
	}
}

func (r *Redis) mGet(args, values []interface{}) ([]interface{}, error) {
	resp := r.pool.Cmd("MGET", args...)
	values = append(values[:0], make([]interface{}, len(args))...)
	if resp.Err != nil {
		return nil, resp.Err
	}
	array, err := resp.Array()
	if err != nil {
		return nil, err
	}
	for i, r := range array {
		if v, err := r.Str(); err == nil {
			values[i] = v
		} else { // if r.IsType(redis.Nil) {
			values[i] = nil
		}
	}
	return values, nil
}

// k/v pair handler
// if v argument in TupleOp is nil then k is absent from db
type TupleOp func(k, v interface{})

// get keys from Scanner, GET them from redis, then
// process them via TupleOp
// if TupleOp returns false: stop and return latest error value
// if error is encountered, finish and return it.
func (r *Redis) Resolve(ctx context.Context, s util.Scanner, fn TupleOp) error {
	ch := make(chan interface{}, queryChannelBuf)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func(ctx context.Context) {
		buf := make([]interface{}, 0, queryChannelBuf)
		values := make([]interface{}, 0, queryChannelBuf)
		ticker := time.NewTicker(queryDrainInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case k := <-ch:
				if buf = append(buf, k); len(buf) < cap(buf) {
					continue
				}
			case <-ticker.C:
				if len(buf) == 0 {
					continue
				}
			}
			values, err := r.mGet(buf, values)
			if err != nil {
				errCh <- err
				return
			}
			for i, key := range buf {
				fn(key, values[i])
			}
			buf = buf[:0]
		}
	}(ctx)

	for s.HasNext() {
		select {
		case <-ctx.Done():
			return s.Err()
		case ch <- s.Next():
		case err := <-errCh:
			return err
		}
	}

	return s.Err()
}

func (r *Redis) ResolveScan(ctx context.Context, fn TupleOp) error {
	return r.Resolve(ctx, util.NewScanner(r.pool,
		util.ScanOpts{Command: "SCAN", Count: scanCount}), fn)
}

func (r *Redis) ResolveEvents(ctx context.Context, fn TupleOp) error {
	return r.Resolve(ctx, r.NewKeyEventSource(), fn)
}
