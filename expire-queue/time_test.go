package eq

import (
	"testing"
	"time"
)

func benchmarkTime(b *testing.B, res time.Duration) {
	ts := NewTimeChanSource(res)

	for i := 0; i < b.N; i++ {
		ts.Now()
	}
}

func BenchmarkTimeChanNow(b *testing.B) {
	benchmarkTime(b, 0)
}

func BenchmarkTimeChan1us(b *testing.B) {
	benchmarkTime(b, time.Microsecond)
}

func BenchmarkTimeChan1ms(b *testing.B) {
	benchmarkTime(b, time.Millisecond)
}

func BenchmarkTimeChan100ms(b *testing.B) {
	benchmarkTime(b, 100*time.Millisecond)
}

func BenchmarkTimeChan1s(b *testing.B) {
	benchmarkTime(b, time.Second)
}

func benchmarkDefer(b *testing.B, n int) {
	ts := &TimeDeferSource{N: n}

	for i := 0; i < b.N; i++ {
		ts.Now()
	}
}

func BenchmarkTimeDefer10(b *testing.B) {
	benchmarkDefer(b, 10)
}

func BenchmarkTimeDefer100(b *testing.B) {
	benchmarkDefer(b, 100)
}

func BenchmarkTimeDefer10000(b *testing.B) {
	benchmarkDefer(b, 10000)
}

func BenchmarkTimeDefer1000000(b *testing.B) {
	benchmarkDefer(b, 1000000)
}
