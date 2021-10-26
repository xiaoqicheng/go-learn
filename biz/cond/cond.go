package cond

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

/**
 @desc cond原语
 @date 2021-10-25
 @author cf
 */


//Go 标准库提供 Cond 原语的目的是，为等待/通知场景下的并发问题提供支持。Cond 通常应用于等待某个条件的一组 goroutine，等条件变为 true 的时候，其中一个 goroutine 或者所有的 goroutine 都会被唤醒执行。

//那这里等待的条件是什么呢？等待的条件，可以是某个变量达到了某个阈值或者某个时间点，也可以是一组变量分别都达到了某个阈值，还可以是某个对象的状态满足了特定的条件。总结来讲，等待的条件是一种可以用来计算结果是 true 还是 false 的条件。


//example

func ExampleCond() {
	c := sync.NewCond(&sync.Mutex{})

	var ready int

	for i := 0; i < 10; i++ {
		go func(i int) {
			time.Sleep(time.Duration(rand.Int63n(10)) * time.Second)

			// 加锁更改等待条件
			c.L.Lock()
			ready++
			c.L.Unlock()

			log.Printf("运动员#%d 已准备就绪\n", i)

			// 广播唤醒所有的等待者
			c.Broadcast()
		}(i)
	}

	c.L.Lock()
	for ready != 10 { //检查10次
		c.Wait() //必须加锁
		log.Println("裁判被唤醒一次")
	}
	c.L.Unlock()

	//所有的运动员是否就绪
	log.Println("所有的运动员是否准备就绪，比赛开始。。。。。")
}

//你看，Cond 的使用其实没那么简单。它的复杂在于：一，这段代码有时候需要加锁，有时候可以不加；二，Wait 唤醒后需要检查条件；三，条件变量的更改，其实是需要原子操作或者互斥锁保护的。所以，有的开发者会认为，Cond 是唯一难以掌握的 Go 并发原语。


//常见错误：
//1、 c.wait 不加锁
//2、 只调用了一次wait

//这个东西吧，虽然只有三个基本放大 但是不好用 容易出错，很少使用。目前对我来说仅作为了解
