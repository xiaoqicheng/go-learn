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

/**
 @WaitGroup 常见错误问题
 1、计数器设置为负数，原因有二。一、add(负数) 二、done()次数过多
 2、等所有的add方法调用之后在调用wait,否则可能会导致panic或者不期望的结果
 3、waitGroup 可以重用，但是必须等到上一轮wait完成之后，才能重用waitGroup执行下一轮的add/wait,否则可能会导致panic
 */