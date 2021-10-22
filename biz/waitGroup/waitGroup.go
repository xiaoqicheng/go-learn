package waitGroup

import (
	"fmt"
	"sync"
	"time"
)

//直接上例子

//线程安全计数器
type Counter struct {
	mu sync.Mutex
	count uint64
}

func (c *Counter) Incr() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *Counter) Count() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.count
}

func worker(c *Counter, wg *sync.WaitGroup)  {
	defer wg.Done()
	time.Sleep(time.Second)
	c.Incr()
}

func Example()  {
	var counter Counter
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(&counter, &wg)
	}

	wg.Wait()
	fmt.Println(counter.count)
}