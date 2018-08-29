package lb

type Config struct {
	// default goroutine selector
	Selector
	// Number of worker goroutines
	Workers int
	// Goroutine buffer size, i.e. pending jobs max number
	BufferSize int
}

type LoadBalancer struct {
	Selector
	farm []chan func()
	stop chan bool
}

func New(config Config) *LoadBalancer {
	lb := &LoadBalancer{
		Selector: config.Selector,
		stop:     make(chan bool, config.Workers),
		farm:     make([]chan func(), config.Workers)}

	for i, _ := range lb.farm {
		ch := make(chan func(), config.BufferSize)
		lb.farm[i] = ch
		go func(ch chan func()) {
			for f := range ch {
				f()
			}
			lb.stop <- true
		}(ch)
	}

	return lb
}

// Stop cluster's goroutines.
func (lb *LoadBalancer) Stop() {
	for _, ch := range lb.farm {
		close(ch)
	}
	// wait for goroutines exit
	for _, _ = range lb.farm {
		<-lb.stop
	}
}

// Perform a job on specified `Item`.
func (lb *LoadBalancer) Do(f func()) {
	lb.DoWith(lb.Select(len(lb.farm)), f)
}

// Perform a job on specified `Item` with specific goroutine.
func (lb *LoadBalancer) DoWith(i int, f func()) {
	if N := len(lb.farm); N > 0 {
		lb.farm[i%N] <- f
	}
}

// Perform a bunch of jobs.
func (lb *LoadBalancer) DoBulk(ff []func()) {
	lb.DoBulkWith(lb.Select(len(lb.farm)), ff)
}

// Perform a bunch of jobs with specific goroutine.
func (lb *LoadBalancer) DoBulkWith(i int, ff []func()) {
	lb.DoWith(i, func() {
		for _, x := range ff {
			x()
		}
	})
}

func (lb *LoadBalancer) Workers() int {
	return len(lb.farm)
}
