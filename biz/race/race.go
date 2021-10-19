package race

import (
	"fmt"
	"sync"
)

//Race  Mutex 互斥锁的使用
func Race()  {

	var mu sync.Mutex

	count := 0

	wg := sync.WaitGroup{}

	for i:=0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j :=0; j<10000; j++ {
				//获取锁
				mu.Lock()
				count++
				//释放锁
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Println(count)
}




//Counter  struct 有多个字段 一般会将 Mutex 放在要控制的字段上面，然后使用空格将字段分隔开
type Counter struct {
	counterType int
	name string

	mu sync.Mutex //不需要初始化
	count uint64
}

//RaceStruct struct 结构体使用互斥锁
func RaceStruct()  {
	var counter Counter
	var wg sync.WaitGroup

	for i:=0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j :=0; j<10000; j++ {
				counter.Incr()
			}
		}()
	}
	wg.Wait()

	fmt.Println(counter.Count())
}

//Incr 计数
func (c *Counter) Incr() {
	//获取锁
	c.mu.Lock()
	c.count++
	//释放锁
	c.mu.Unlock()
}

//Count 获取当前计数
func (c *Counter) Count() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}