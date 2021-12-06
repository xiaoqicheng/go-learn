package memoryModel

/**
说到这儿，我想先给你补充三个 Go 语言中和内存模型有关的小知识，掌握了这些，你就能更好地理解下面的内容。
	1.在 Go 语言中，对变量进行零值的初始化就是一个写操作。
	2.如果对超过机器 word（64bit、32bit 或者其它）大小的值进行读写，那么，就可以看作是对拆成 word 大小的几个读写无序进行。
	3.Go 并不提供直接的 CPU 屏障（CPU fence）来提示编译器或者 CPU 保证顺序性，而是使用不同架构的内存屏障指令来实现统一的并发原语。


Go 语言中保证的 happens-before 关系:
除了单个 goroutine 内部提供的 happens-before 保证，Go 语言中还提供了一些其它的 happens-before 关系的保证，下面我来一个一个介绍下。

init 函数
	应用程序的初始化是在单一的 goroutine 执行的。如果包 p 导入了包 q，那么，q 的 init 函数的执行一定 happens before  p 的任何初始化代码。
*** 这里有一个特殊情况需要你记住：main 函数一定在导入的包的 init 函数之后执行
包级别的变量在同一个文件中是按照声明顺序逐个初始化的，除非初始化它的时候依赖其它的变量。同一个包下的多个文件，会按照文件名的排列顺序进行初始化。这个顺序被定义在 Go 语言规范中，而不是 Go 的内存模型规范中。你可以看看下面的例子中各个变量的值：

	var (
		a = c + b
		b = f()
		c = f()
		d = 3
	)

	func f () int {
		d++
		return d
	}

	具体怎么对这些变量进行初始化呢？Go 采用的是依赖分析技术。不过，依赖分析技术保证的顺序只是针对同一包下的变量，而且，只有引用关系是本包变量、函数和非接口的方法，才能保证它们的顺序性。同一个包下可以有多个 init 函数，但是每个文件最多只能有一个 init 函数，多个 init 函数按照它们的文件名顺序逐个初始化。刚刚讲的这些都是不同包的 init 函数执行顺序，下面我举一个具体的例子，把这些内容串起来，你一看就明白了。


首先，我们需要明确一个规则：启动 goroutine 的 go 语句的执行，一定 happens before 此 goroutine 内的代码执行。
根据这个规则，我们就可以知道，如果 go 语句传入的参数是一个函数执行的结果，那么，这个函数一定先于 goroutine 内部的代码被执行。

我们来看一个例子。在下面的代码中，第 8 行 a 的赋值和第 9 行的 go 语句是在同一个 goroutine 中执行的，所以，在主 goroutine 看来，第 8 行肯定 happens before 第 9 行，又由于刚才的保证，第 9 行子 goroutine 的启动 happens before 第 4 行的变量输出，那么，我们就可以推断出，第 8 行 happens before 第 4 行。也就是说，在第 4 行打印 a 的值的时候，肯定会打印出“hello world”。

	var a string

	func f()  {
		print(a)
	}

	func hello()  {
		a = "hello world"
		go f()
	}

刚刚说的是启动 goroutine 的情况，goroutine 退出的时候，是没有任何 happens-before 保证的。所以，如果你想观察某个 goroutine 的执行效果，你需要使用同步机制建立 happens-before 关系，比如 Mutex 或者 Channel。接下来，我会讲 Channel 的 happens-before 的关系保证。

 */

/**
	channel: 通用的 Channel happens-before 关系保证有 4 条规则，我分别来介绍下。
第一条规则是，往 Channel 中的发送操作，happens before 从该 Channel 接收相应数据的动作完成之前，即第 n 个 send 一定 happens before 第 n 个 receive 的完成。
	var ch = make(chan struct{}, 10)
	var s string

	func f()  {
		s = "hello,world"
		ch <- struct{}{}
	}

	func exampleTest()  {
		go f()
		<-ch
		print(s)
	}

第二条规则是，close 一个 Channel 的调用，肯定 happens before 从关闭的 Channel 中读取出一个零值。

第三条规则是，对于 unbuffered 的 Channel，也就是容量是 0 的 Channel，从此 Channel 中读取数据的调用一定 happens before 往此 Channel 发送数据的调用完成。
所以，在上面的这个例子中呢，如果想保持同样的执行顺序，也可以写成这样：
	var ch = make(chan struct{})
	var s string

	func f()  {
		s = "hello,world"
		<-ch
	}

	func exampleTest()  {
		go f()
		ch <- struct{}{}
		print(s)
	}

第四条规则是，如果 Channel 的容量是 m（m>0），那么，第 n 个 receive 一定 happens before 第 n+m 个 send 的完成。
 */

/**
Mutex/RWMutex:
	1.第 n 次的 m.Unlock 一定 happens before 第 n+1 m.Lock 方法的返回；
	2.对于读写锁 RWMutex m，如果它的第 n 个 m.Lock 方法的调用已返回，那么它的第 n 个 m.Unlock 的方法调用一定 happens before 任何一个 m.RLock 方法调用的返回，只要这些 m.RLock 方法调用 happens after 第 n 次 m.Lock 的调用的返回。这就可以保证，只有释放了持有的写锁，那些等待的读请求才能请求到读锁。
	3.对于读写锁 RWMutex m，如果它的第 n 个 m.RLock 方法的调用已返回，那么它的第 k （k<=n）个成功的 m.RUnlock 方法的返回一定 happens before 任意的 m.RUnlockLock 方法调用，只要这些 m.Lock 方法调用 happens after 第 n 次 m.RLock。

	var mu sync.Mutex
	var s string
	func foo()  {
		s = "hello,world"
		mu.Unlock()
	}
	func ExampleTest()  {
		mu.Lock()
		go foo()
		mu.Lock()
		print(s)
	}
 */

/**
	WaitGroup:   Wait 方法等到计数值归零之后才返回
对于一个 WaitGroup 实例 wg，在某个时刻 t0 时，它的计数值已经不是零了，假如 t0 时刻之后调用了一系列的 wg.Add(n) 或者 wg.Done()，并且只有最后一次调用 wg 的计数值变为了 0，那么，可以保证这些 wg.Add 或者 wg.Done() 一定 happens before t0 时刻之后调用的 wg.Wait 方法的返回。

 */

/**
	Once: 对于 once.Do(f) 调用，f 函数的那个单次调用一定 happens before 任何 once.Do(f) 调用的返回。换句话说，就是函数 f 一定会在 Do 方法返回之前执行。

	var once sync.Once
	var s string
	func foo()  {
		s = "hello,world"
	}
	func ExampleTest()  {
		once.Do(foo)
		print(s)
	}
 */


// 总结图：../images/memoryModel.jpg