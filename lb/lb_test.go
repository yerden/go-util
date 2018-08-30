package lb

import (
	"github.com/yerden/go-util/assert"
	"hash/crc32"
	"math/rand"
	//"strconv"
	"testing"
	"time"
)

func TestLB1(t *testing.T) {
	lb := New(Config{
		Workers:    5,
		Selector:   NewSelector(RoundRobin),
		BufferSize: 32})
	defer lb.Stop()
	a := assert.New(t)
	a.NotNil(lb)

	a.Equal(lb.Workers(), 5)

	ch := make(chan int)

	lb.Do(func() {
		ch <- 1
	})

	a.Equal(<-ch, 1)
	lb.DoWith(2, func() {
		ch <- 2
	})

	a.Equal(<-ch, 2)
}

func TestLB2(t *testing.T) {
	a := assert.New(t)
	var sum, sum1 int
	workers := 5
	ch := make(chan int, workers)
	// test data
	data := make([]int, 100)
	for i, _ := range data {
		data[i] = rand.Int()
	}

	// direct sum
	for _, i := range data {
		sum += i
	}

	// do it with LoadBalancer
	getF := func(pos, off int) func() {
		return func() {
			var s int
			for _, i := range data[pos : pos+off] {
				s += i
			}
			ch <- s
		}
	}

	lb := New(Config{
		Workers:    workers,
		Selector:   NewSelector(RoundRobin),
		BufferSize: 32})
	for i := 0; i < workers; i++ {
		lb.Do(getF(20*i, 20))
	}
	for i := 0; i < workers; i++ {
		sum1 += <-ch
	}

	a.Equal(sum, sum1)
}

type myjob func()

func (j myjob) Do() {
	j()
}

func TestLB3(t *testing.T) {
	a := assert.New(t)
	var sum, sum1 int
	workers := 5
	ch := make(chan int, workers)
	// test data
	data := make([]int, 100)
	for i, _ := range data {
		data[i] = rand.Int()
	}

	// direct sum
	for _, i := range data {
		sum += i
	}

	// do it with LoadBalancer
	getF := func(pos, off int) func() {
		return func() {
			var s int
			for _, i := range data[pos : pos+off] {
				s += i
			}
			ch <- s
		}
	}

	team := NewTeam(TeamConfig{
		Number:   workers,
		Selector: NewSelector(RoundRobin),
		WorkerConfig: WorkerConfig{
			BacklogSize:   32,
			ChannelBuffer: 32}})
	for i := 0; i < workers; i++ {
		team.Push(myjob(getF(20*i, 20)))
	}
	team.Close()
	for i := 0; i < workers; i++ {
		sum1 += <-ch
	}

	a.Equal(sum, sum1)
}

type myint int

func (x *myint) Do() {
	*x = *x + 1
}

func TestWorker(t *testing.T) {
	a := assert.New(t)

	w := NewWorker(WorkerConfig{BacklogSize: 14, ChannelBuffer: 2})
	a.NotNil(w)

	d := make([]myint, 20)

	for i, _ := range d {
		d[i] = myint(i)
		w.Push(&d[i])
	}
	w.Close()
	for i, _ := range d {
		a.Equal(i+1, int(d[i]))
	}
}

type chunk struct {
	data   []byte
	table  *crc32.Table
	output chan uint32
}

func (c *chunk) Do() {
	c.output <- crc32.Update(0, c.table, c.data)
}

func prepareChunks(num, chunkSize int) ([]chunk, chan uint32) {
	ch := make(chan uint32, num+10)
	data := make([]byte, num*chunkSize)
	chunks := make([]chunk, num)
	rand.Seed(time.Now().Unix())
	rand.Read(data)

	for i, _ := range chunks {
		chunks[i].data = data[i*chunkSize : (i+1)*chunkSize]
		chunks[i].table = crc32.MakeTable(crc32.IEEE)
		chunks[i].output = ch
	}
	return chunks, ch
}

func computeSerial(num, chunkSize int) func() {
	chunks, ch := prepareChunks(num, chunkSize)

	return func() {
		res := make([]uint32, num)
		for i, _ := range chunks {
			c := &chunks[i]
			c.Do()
		}
		for i, _ := range chunks {
			res[i] = <-ch
		}
	}
}

func computeWithLB(num, chunkSize, workers, buffer int) func() {
	chunks, ch := prepareChunks(num, chunkSize)
	getF := func(c *chunk) func() {
		return c.Do
	}

	return func() {
		lb := New(Config{
			Workers:    workers,
			Selector:   NewSelector(RoundRobin),
			BufferSize: buffer})
		defer lb.Stop()

		res := make([]uint32, num)
		for i, _ := range chunks {
			c := &chunks[i]
			lb.Do(getF(c))
		}
		for i, _ := range chunks {
			res[i] = <-ch
		}
	}
}

func computeWithTeam(num, chunkSize, workers, buffer int) func() {
	chunks, ch := prepareChunks(num, chunkSize)

	return func() {
		team := NewTeam(TeamConfig{
			Number:   workers,
			Selector: NewSelector(RoundRobin),
			WorkerConfig: WorkerConfig{
				BacklogSize:   buffer,
				ChannelBuffer: buffer}})
		res := make([]uint32, num)
		for i, _ := range chunks {
			c := &chunks[i]
			team.Push(c)
		}
		team.Close()
		for i, _ := range chunks {
			res[i] = <-ch
		}
	}
}

func BenchmarkWithTeam(b *testing.B) {
	f := computeWithTeam(1000, 1000, 12, 1000)

	for i := 0; i < b.N; i++ {
		f()
	}
}

func BenchmarkWithLB(b *testing.B) {
	f := computeWithLB(1000, 1000, 12, 10)

	for i := 0; i < b.N; i++ {
		f()
	}
}

func BenchmarkSerial(b *testing.B) {
	f := computeSerial(1000, 1000)

	for i := 0; i < b.N; i++ {
		f()
	}
}
