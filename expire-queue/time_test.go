package eq

import (
	"testing"
	"time"
)

func benchmarkTime(b *testing.B, res time.Duration) {
	ts := NewTimeSource(res)

	for i := 0; i < b.N; i++ {
		ts.Now()
	}
}

func BenchmarkTimeNow(b *testing.B) {
	benchmarkTime(b, 0)
}

func BenchmarkTime1us(b *testing.B) {
	benchmarkTime(b, time.Microsecond)
}

func BenchmarkTime1ms(b *testing.B) {
	benchmarkTime(b, time.Millisecond)
}

func BenchmarkTime100ms(b *testing.B) {
	benchmarkTime(b, 100*time.Millisecond)
}

func BenchmarkTime1s(b *testing.B) {
	benchmarkTime(b, time.Second)
}
