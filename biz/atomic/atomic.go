package atomic

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

/**
	@desc 保证原子性操作
	@date 2021-11-18
	@author cf
 */

//Package sync/atomic 实现了同步算法底层的原子的内存操作原语，我们把它叫做原子操作原语，它提供了一些实现原子操作的方法。之所以叫原子操作，是因为一个原子在执行的时候，其它线程不会看到执行一半的操作结果。在其它线程看来，原子操作要么执行完了，要么还没有执行，就像一个最小的粒子 - 原子一样，不可分割。   因为不用的cpu架构原子的操作指令不用，所以一定要使用atomic提供的方法

//atomic 原子操作的应用场景

//1、问题不涉及到对资源复杂的竞争逻辑，只是会并发地读写这个标志，这类场景就适合使用 atomic 的原子操作。具体怎么做呢？你可以使用一个 uint32 类型的变量，如果这个变量的值是 0，就标识没有任务在执行，如果它的值是 1，就标识已经有任务在完成了。你看，是不是很简单呢？

//2、atomic 原子操作还是实现 lock-free 数据结构的基石。在实现 lock-free 的数据结构时，我们可以不使用互斥锁，这样就不会让线程因为等待互斥锁而阻塞休眠，而是让线程保持继续处理的状态。另外，不使用互斥锁的话，lock-free 的数据结构还可以提供并发的性能。 lock-free 数据结构实现起来比较复杂，目前本人没兴趣知道，有兴趣的自己查资料

//atomic 提供的方法

// * 注意 * ：atomic 操作的对象是一个地址，你需要把可寻址的变量的地址作为参数传递给方法，而不是把变量的值传递给方法

//1、 Add方法 给地址个参数地址中的值增加一个delta值  func AddInt32(addr *int32, delta int32) (new int32)
//2、 CAS (CompareAndSwap)方法 判断当前addr 地址里的值是不是old 如果不是返回false，如果是 把当前地址的值替换为 new 并返回true func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)

//3、Swap 方法 无须比较粗暴替换 func SwapInt32(addr *int32, new int32) (old int32)

//4、Load 方法 Load 方法会取出 addr 地址中的值，即使在多处理器、多核、有 CPU cache 的情况下，这个操作也能保证 Load 是一个原子操作。func LoadInt32(addr *int32) (val int32)

//5、Store() Store 方法会把一个值存入到指定的 addr 地址中，即使在多处理器、多核、有 CPU cache 的情况下，这个操作也能保证 Store 是一个原子操作。别的 goroutine 通过 Load 读取出来，不会看到存取了一半的值。func StoreInt32(addr *int32, val int32)

//6、Value 类型 举例说明
type Config struct {
	NodeName string
	Addr	 string
	Count 	 int32
}

func loadNewConfig() Config {
	return Config{
		NodeName: "北京",
		Addr: "10.77.95.27",
		Count: rand.Int31(),
	}
}

func ExampleAtomicValue()  {
	var config atomic.Value
	config.Store(loadNewConfig())
	var cond = sync.NewCond(&sync.Mutex{})

	//设置新的config
	go func() {
		for  {
			time.Sleep(time.Duration(5+rand.Int31n(5)) * time.Second)
			config.Store(loadNewConfig())
			cond.Broadcast() //通知等待这配置已变更
		}
	}()

	go func() {
		for  {
			cond.L.Lock()
			cond.Wait()		//等待变更信息
			c := config.Load().(Config) //读取新的配置
			fmt.Printf("new config: %+v\n", c)
			cond.L.Unlock()
		}
	}()

	select {}
}


//不过有一点让人觉得不爽的是，或者是让熟悉面向对象编程的程序员不爽的是，函数调用有一点点麻烦。所以，有些人就对这些函数做了进一步的包装，跟 atomic 中的 Value 类型类似，这些类型也提供了面向对象的使用方式，比如关注度比较高的 uber-go/atomic，它定义和封装了几种与常见类型相对应的原子操作类型，这些类型提供了原子操作的方法。这些类型包括 Bool、Duration、Error、Float64、Int32、Int64、String、Uint32、Uint64 等。
