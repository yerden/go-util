package eq

import (
	"runtime"
	"sync/atomic"
	"time"
)

type TimeSource struct {
	res  time.Duration
	nsec int64
	done chan bool
}

func NewTimeSource(res time.Duration) *TimeSource {
	ts := &TimeSource{}
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

	runtime.SetFinalizer(ts, func(ts *TimeSource) { close(ts.done) })
	return ts
}

func (ts *TimeSource) Now() time.Time {
	if ts.res == 0 {
		return time.Now()
	}

	return time.Unix(0, atomic.LoadInt64(&ts.nsec))
}

func (ts *TimeSource) UnixNano() int64 {
	if ts.res == 0 {
		return time.Now().UnixNano()
	}

	return atomic.LoadInt64(&ts.nsec)
}
