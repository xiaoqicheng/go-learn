package readersWriters

import (
	"fmt"
	"sync"
	"time"
)

/*
 @desc 读写锁 在某一时刻只能由任意数量的reader持有，或者是只被单个writer持有
 @date 2021-10-19
 @author cf 
 */


//example：计数器的读写操作

//一个线程安全的计数器
type Counter struct {
	mu sync.RWMutex
	count uint64
}

// 计数器例子
func counterExample()  {
	var counter Counter

	for i := 0; i < 10; i++ {
		go func() {
			for  {
				counter.Count() //计数器读取
				time.Sleep(time.Millisecond)
			}
		}()
	}

	for  {
		counter.Incr() //写入
		time.Sleep(time.Second)
	}
}

//计数
func (c *Counter) Incr()  {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

//读取
func (c *Counter) Count() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.count
}

/*
	readers-writers 读写优先级问题。三种情况，1、读优先 2、写优先 3、随便抢 who care
	1、Read-preferring: 读优先的设计可以提供很高的并发性，但是可能会导致饥饿写
	2、Write-preferring:写优先优先保证了写操作，避免了饥饿写。但是可能会导致饥饿读
	3、不指定优先级：不指定优先级，一视同仁，解决了饥饿问题
 */

//踩坑点
//1、不可复制  与Mutex的不可复制一样。可以使用 vet 工具在变量赋值，函数传参，函数返回值，遍历数据，struct初始化等，检查是否有读写隐士复制的情景，example：go vet main.go

//2、重入导致死锁

// example1: 读写锁内部是实现使用Mutex对writer的并发访问，因为Mutex有重入问题

func exampleMain()  {
	l := &sync.RWMutex{}
	foo(l)
}

func foo(l *sync.RWMutex)  {
	fmt.Println("in foo")
	l.Lock()
	bar(l)
	l.Unlock()
}

func bar(l *sync.RWMutex)  {
	l.Lock()
	fmt.Println("in bar")
	l.Unlock()
}

// example2: reader活跃的时候 writer会等待，如果在reader的时候调用写的操作就会造成死锁

// example3: 文字太多懒得打，请看 images文件夹中文件名为 remutex_race_example3.jpg,代码举例如下

func Example3()  {
	var mu sync.RWMutex

	//writer 稍微等待，然后制造一个调用Lock的场景
	go func() {
		time.Sleep(200 * time.Millisecond)
		mu.Lock()
		fmt.Println("Lock")
		time.Sleep(100 * time.Millisecond)
		mu.Unlock()
		fmt.Println("unlock")
	}()

	go func() {
		factorial(&mu, 10) //计算10的阶乘 10！
	}()

	select {}
}

func factorial(m *sync.RWMutex, n int) int {
	if n < 1 {
		return 0
	}
	fmt.Println("RLock")
	m.RLock()
	defer func() {
		fmt.Println("RUnlock")
		m.RUnlock()
	}()

	time.Sleep(100 * time.Millisecond)
	return factorial(m, n-1)*n //递归调用
}


//3、释放未加锁的RWMutex，和 Mutex一样 未成对出现