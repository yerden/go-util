package eq

import (
	"runtime"
	"sync/atomic"
	"time"
)

type TimeSource interface {
	Now() time.Time
}

type TimeChanSource struct {
	res  time.Duration
	nsec int64
	done chan bool
}

func NewTimeChanSource(res time.Duration) *TimeChanSource {
	ts := &TimeChanSource{}
	if res == 0 {
		return ts
	}

	ts.res = res
	ts.done = make(chan bool)
	go func() {
		ticker := time.NewTicker(res)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				atomic.StoreInt64(&ts.nsec, t.UnixNano())
			case <-ts.done:
				return
			}
		}
	}()

	runtime.SetFinalizer(ts, func(ts *TimeChanSource) { close(ts.done) })
	return ts
}

func (ts *TimeChanSource) Now() time.Time {
	if ts.res == 0 {
		return time.Now()
	}

	return time.Unix(0, atomic.LoadInt64(&ts.nsec))
}

type TimeDeferSource struct {
	N, i int
	t    time.Time
}

func (ts *TimeDeferSource) Now() time.Time {
	if ts.i%ts.N == 0 {
		ts.i = 0
		ts.t = time.Now()
	}
	return ts.t
}
