package redis

import (
	"context"
	"fmt"
	"github.com/mediocregopher/radix/v3"
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
	readTimeout        = 10 * time.Second
	queryChannelBuf    = 128
	queryDrainInterval = 100 * time.Millisecond
	scanCount          = 200
)

type Redis struct {
	pool *radix.Pool
	conf RedisConfig
}

func NewRedis(c RedisConfig) (*Redis, error) {
	connfn := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialSelectDB(c.DbIndex),
			radix.DialReadTimeout(readTimeout))
	}
	pool, err := radix.NewPool(c.Network, c.Addr, 10, radix.PoolConnFunc(connfn))
	return &Redis{pool: pool, conf: c}, err
}

func (r *Redis) Get(key string) (string, error) {
	var val string
	return val, r.pool.Do(radix.Cmd(&val, "GET", key))
}

func getEvent(channel string) string {
	if s := strings.SplitN(channel, ":", 2); len(s) > 1 {
		return s[1]
	}
	return ""
}

// k/v pair handler
// if v argument in TupleOp is nil then k is absent from db
type TupleOp func(k, v interface{})

// receive PubSubMessages via channel, evaluate in Redis and
// feed to handler
func (r *Redis) consume(ctx context.Context, msgCh <-chan radix.PubSubMessage, fn TupleOp) error {
	args := make([]string, 0, queryChannelBuf)
	values := make([]radix.MaybeNil, len(args))

	for i, _ := range values {
		values[i].Rcv = new(string)
	}

	ticker := time.NewTicker(queryDrainInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-msgCh:
			if msg.Message == nil {
				// default message, channel is closing
				return nil
			}
			if args = append(args, string(msg.Message)); len(args) < cap(args) {
				continue
			}
		case <-ticker.C:
			if len(args) == 0 {
				continue
			}
		}

		if err := r.pool.Do(radix.Cmd(&values, "MGET", args...)); err != nil {
			return err
		}
		for i, v := range values {
			if v.Nil {
				fn(args[i], nil)
			} else {
				fn(args[i], *(v.Rcv.(*string)))
			}
		}
		args = args[:0]
	}

	return nil
}

func (r *Redis) ConsumeKeyEvents(ctx context.Context, fn TupleOp) error {
	msgCh := make(chan radix.PubSubMessage, queryChannelBuf)
	defer close(msgCh)

	// subscribe
	ps := radix.PersistentPubSub(r.conf.Network, r.conf.Addr,
		func(network, addr string) (radix.Conn, error) {
			return radix.Dial(network, addr, radix.DialReadTimeout(readTimeout))
		})
	defer ps.Close()
	ps.PSubscribe(msgCh, fmt.Sprintf("__keyevent@%d__:*", r.conf.DbIndex))

	return r.consume(ctx, msgCh, fn)
}

func (r *Redis) ConsumeScan(ctx context.Context, fn TupleOp) error {
	scanner := radix.NewScanner(r.pool, radix.ScanOpts{Command: "SCAN", Count: scanCount})
	errCh := make(chan error, 1)
	msgCh := make(chan radix.PubSubMessage, queryChannelBuf)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func(ctx context.Context) { errCh <- r.consume(ctx, msgCh, fn) }(ctx)
	var key string
	for scanner.Next(&key) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case msgCh <- radix.PubSubMessage{Message: []byte(key)}:
		}
	}

	return scanner.Close()
}
