package lb

import (
	//"fmt"
	"sync"
)

type Job interface {
	Do()
}

type Worker interface {
	Push(Job)
	Close()
}

// Configuration of a worker
type WorkerConfig struct {
	// size of backlog for incoming jobs
	BacklogSize int
	// buffer of channel to hold backlog
	ChannelBuffer int
}

type TeamConfig struct {
	Selector
	WorkerConfig
	Number int
}

type worker struct {
	pool    *sync.Pool
	buffer  []Job
	jobsCh  chan []Job
	closeCh chan bool // closing channel
	size    int
}

type team struct {
	Selector
	workers []Worker
}

func NewWorker(c WorkerConfig) *worker {
	prealloc := c.BacklogSize
	if prealloc <= 0 {
		prealloc = 1
	}
	pool := &sync.Pool{New: func() interface{} {
		return make([]Job, 0, prealloc)
	}}
	w := &worker{
		size:    c.BacklogSize,
		pool:    pool,
		buffer:  pool.Get().([]Job),
		jobsCh:  make(chan []Job, c.ChannelBuffer),
		closeCh: make(chan bool),
	}
	go w.handle()
	return w
}

func (w *worker) execute() {
	w.jobsCh <- w.buffer
}

func (w *worker) Push(j Job) {
	if w.buffer = append(w.buffer, j); len(w.buffer) >= w.size {
		w.execute()
		w.buffer = w.pool.Get().([]Job)
	}
}

func (w *worker) handle() {
	defer close(w.closeCh)
	for buffer := range w.jobsCh {
		for _, j := range buffer {
			j.Do()
		}
		w.pool.Put(buffer[:0])
	}
}

// Close() processes the remainder of jobs
// and closes job channel
func (w *worker) Close() {
	w.execute()
	close(w.jobsCh)
	<-w.closeCh
}

func NewTeam(c TeamConfig) *team {
	workers := make([]Worker, c.Number)
	for i, _ := range workers {
		workers[i] = NewWorker(c.WorkerConfig)
	}
	return &team{c.Selector, workers}
}

func (t *team) Push(j Job) {
	id := t.Select(len(t.workers))
	t.workers[id].Push(j)
}

func (t *team) Close() {
	for i, _ := range t.workers {
		t.workers[i].Close()
	}
}
