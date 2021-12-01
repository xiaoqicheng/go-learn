package channel

import (
	"fmt"
	"reflect"
	"time"
)

/**
	@desc 典型应用模式
	@date 2021-11-19
	@author cf
 */

//使用反射 结合 select 处理 不定数量的chan

func ExampleSelectChannel()  {
	var ch1 = make(chan int, 10)
	var ch2 = make(chan int, 10)

	//创建SelectCase
	var cases = createCases(ch1, ch2)

	//执行10次select
	for i := 0; i < 10; i++ {
		chosen, recv, ok := reflect.Select(cases)
		if recv.IsValid() {
			fmt.Println("recv:", cases[chosen].Dir, recv, ok)
		}else {
			fmt.Println("send:", cases[chosen].Dir, ok)
		}
	}
}

func createCases(chs ...chan int) []reflect.SelectCase {
	var cases []reflect.SelectCase

	// 创建recv case
	for _, ch := range chs {
		cases = append(cases, reflect.SelectCase{
			Dir: reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}

	//创建send case
	for i, ch := range chs {
		v := reflect.ValueOf(i)
		cases = append(cases, reflect.SelectCase{
			Dir: reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: v,
		})
	}

	return cases
}

//============典型的应用场景==================
/*(
	1、信息交流 :从 chan 的内部实现看，它是以一个循环队列的方式存放数据，所以，它有时候也会被当成线程安全的队列和 buffer 使用。一个 goroutine 可以安全地往 Channel 中塞数据，另外一个 goroutine 可以安全地从 Channel 中读取数据，goroutine 就可以安全地实现信息交流了。

	2、数据传递 ：这类场景有一个特点，就是当前持有数据的 goroutine 都有一个信箱，信箱使用 chan 实现，goroutine 只需要关注自己的信箱中的数据，处理完毕后，就把结果发送到下一家的信箱中。
		type Token struct {

		}

		func newWork(id int, ch, nextCh chan Token)  {
			for  {
				token := <- ch   //获取令牌
				fmt.Println((id + 1)) //id 从 1 开始
				time.Sleep(time.Second)
				nextCh <- token
			}
		}

		func exampleTwo()  {
			chs := []chan Token{make(chan Token), make(chan Token), make(chan Token), make(chan Token)}

			for i :=0; i < 4; i++ {
				go newWork(i, chs[i], chs[(i+1)%4])
			}

			//首先把令牌交给第一个worker
			chs[0] <- struct{}{}

			select {

			}
		}

	3、信号通知 ：chan 类型有这样一个特点：chan 如果为空，那么，receiver 接收数据的时候就会阻塞等待，直到 chan 被关闭或者有新的数据到来。利用这个机制，我们可以实现 wait/notify 的设计模式。===== 除了正常的业务处理时的 wait/notify，我们经常碰到的一个场景，就是程序关闭的时候，我们需要在退出之前做一些清理（doCleanup 方法）的动作。这个时候，我们经常要使用 chan。

	func exampleThree()  {
		var closing = make(chan struct{})
		var closed = make(chan struct{})

		go func() {
			//模拟业务处理

			for  {
				select {
				case <- closing:
					return
				default:
					// 业务计算
					time.Sleep(100 * time.Millisecond)

				}
			}
		}()

		// 处理 CTRL+C等中断信号
		termChan := make(chan os.Signal)
		signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
		<- termChan
		close(closing)

		//执行退出之前的清理动作
		go doCleanup(closed)

		select {
		case <-closed:
		case <-time.After(time.Second):
			fmt.Println("清理超时，不等了")
		}

		fmt.Println("优雅退出")
	}

	func doCleanup(closed chan struct{})  {
		time.Sleep(time.Minute)
		close(closed)
	}

	4、锁 ：使用 chan 也可以实现互斥锁。 要想使用 chan 实现互斥锁，至少有两种方式。一种方式是先初始化一个 capacity 等于 1 的 Channel，然后再放入一个元素。这个元素就代表锁，谁取得了这个元素，就相当于获取了这把锁。另一种方式是，先初始化一个 capacity 等于 1 的 Channel，它的“空槽”代表锁，谁能成功地把元素发送到这个 Channel，谁就获取了这把锁。

	//使用chan实现互斥锁
	type Mutex struct {
		ch chan struct{}
	}

	//使用锁需要初始化
	func NewMutex() *Mutex {
		mu := &Mutex{make(chan struct{}, 1)}
		mu.ch <- struct{}{}
		return mu
	}

	//请求锁，直到获取到
	func (m *Mutex) Lock()  {
		<- m.ch
	}

	//解锁
	func (m *Mutex) UnLock() {
		select {
		case m.ch <- struct{}{}:
		default:
			panic("unlock of unlocked mutex")
		}
	}

	//尝试获取锁
	func (m *Mutex) TryLock() bool {
		select {
		case <- m.ch :
			return true
		default:

		}

		return false
	}

	//加入一个超时的设置
	func (m *Mutex) LockTimeout(timeout time.Duration) bool {
		timer := time.NewTicker(timeout)
		select {
		case <-m.ch:
			timer.Stop()
			return true
		case <-timer.C:

		}
		return false
	}

	//锁是否已被持有
	func (m *Mutex) IsLocked() bool  {
		return len(m.ch) == 0
	}

	func example()  {
		m := NewMutex()
		ok := m.TryLock()
		fmt.Printf("locked v %v\n", ok)
		ok = m.TryLock()
		fmt.Printf("locked %v\n", ok)
	}

 */


//任务编排：接下来，我来重点介绍下多个 chan 的编排方式，总共 5 种，分别是 Or-Done 模式、扇入模式、扇出模式、Stream 和 Map-Reduce。
/**
	or-done 模式 : 信号通知模式；比如，你发送同一个请求到多个微服务节点，只要任意一个微服务节点返回结果，就算成功，这个时候，就可以参考下面的实现：
 */

func or(channels ...<-chan interface{}) <- chan interface{} {
	//特殊情况 只有一个或者 0 个chan
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan interface{})
	go func() {
		defer close(orDone)
		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default: //超过两个 递归操作
			m := len(channels)/2
			select {
			case <-or(channels[:m]...):
			case <-or(channels[m:]...):
			}
		}
	}()

	return orDone
}

//测试程序
func sig(after time.Duration) <-chan interface{}{
	c := make(chan interface{})
	go func() {
		defer close(c)
		time.Sleep(after)
	}()

	return c
}

func TestOr()  {
	start := time.Now()
	<-or( //当 chan 的数量大于 2 时，使用递归的方式等待信号
		sig(10*time.Second),
		sig(20*time.Second),
		sig(30*time.Second),
		sig(40*time.Second),
		sig(50*time.Second),
		sig(00*time.Second),
		)
	fmt.Printf("done after %v", time.Since(start))
}

//chan太多 递归不是一个很好的解决方式，可以使用反射的方法
func orReflect(channels ...<-chan interface{}) <-chan interface{} {
	//特殊情况 只有一个或者 0 个chan
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan interface{})
	go func() {
		defer close(orDone)
		//利用反射构建SelectCase
		var cases [] reflect.SelectCase
		for _, c := range channels{
			cases = append(cases, reflect.SelectCase{
				Dir: reflect.SelectRecv,
				Chan: reflect.ValueOf(c),
			})
		}

		//随机选择一个可用的case
		reflect.Select(cases)
	}()

	return orDone
}

/**
	扇入模式：在软件工程中，模块的扇入是指有多少个上级模块调用它。而对于我们这里的 Channel 扇入模式来说，就是指有多个源 Channel 输入、一个目的 Channel 输出的情况。扇入比就是源 Channel 数量比 1。
 */
//反射的代码比较简短，易于理解，主要就是构造出 SelectCase slice，然后传递给 reflect.Select 语句。
func fanInReflect(chans ...<-chan interface{}) <-chan interface{}{
	out := make(chan interface{})
	go func() {
		defer close(out)

		//构造SelectCase slice
		//var cases []reflect.SelectCase
		cases := make([]reflect.SelectCase, 0, len(chans))
		for _, c := range chans{
			cases = append(cases, reflect.SelectCase{
				Dir: reflect.SelectRecv,
				Chan: reflect.ValueOf(c),
			})
		}

		//循环 从 cases中选择一个可用的
		for len(chans) > 0 {
			i, v, ok := reflect.Select(cases)
			if !ok { //此channel已经close
				cases = append(cases[:i], cases[i+1:]...)
				continue
			}
			out <- v.Interface()
		}
	}()

	return out
}

//递归模式
func fanInRec(chans ...<-chan interface{}) <-chan interface{}{
	switch len(chans) {
	case 0:
		c := make(chan interface{})
		close(c)
		return c
	case 1:
		return chans[0]
	case 2:
		return mergeTwo(chans[0], chans[1])
	default:
		m := len(chans) / 2
		return mergeTwo(fanInRec(chans[:m]...), fanInRec(chans[m+1:]...))
	}
}

func mergeTwo(a, b <-chan interface{}) <-chan interface{}{
	c := make(chan interface{})
	go func() {
		defer close(c)
		for a != nil || b != nil { //只要还有可读的chan
			select {
			case v, ok := <-a:
				if !ok { // a 已关闭，设置为 nil
					a = nil
					continue
				}
				c <- v
			case v, ok := <-b:
				if !ok { // b 已关闭，设置为nil
					b = nil
					continue
				}
				c <- v
			}
		}
	}()

	return c
}

/**
	扇出模式：扇出模式和扇入模式是相反的；扇出模式只有一个输入源 Channel，有多个目标 Channel，扇出比就是 1 比目标 Channel 数的值，经常用在设计模式中的中（观察者设计模式定义了对象间的一种一对多的组合关系。这样一来，一个对象的状态发生变化时，所有依赖于它的对象都会得到通知并自动刷新）。在观察者模式中，数据变动后，多个观察者都会收到这个变更信号。
 */

//你也可以尝试使用反射的方式来实现，我就不列相关代码了，希望你课后可以自己思考下。
func fanOut(ch <-chan interface{}, out []chan interface{}, async bool)  {
	go func() {
		defer func() { //退出时关闭所有的输出chan
			for i:=0; i< len(out); i++ {
				close(out[i])
			}
		}()

		for v := range ch { //从输出chan中读取数据
			v := v
			for i := 0; i < len(out); i++ {
				i:=i
				if async {
					go func() {
						out[i] <- v //放入到输出chan中，异步方式
					}()
				}else {
					out[i] <- v //放入到输出chan中，同步方式
				}
			}
		}
	}()
}

/**
	Stream: 这里我来介绍一种把 Channel 当作流式管道使用的方式，也就是把 Channel 看作流（Stream），提供跳过几个元素，或者是只取其中的几个元素等方法。
 */

//首先，我们提供创建流的方法。这个方法把一个数据 slice 转换成流：
func asStream(done <-chan struct{}, values ...interface{}) <-chan interface{}{
	s := make(chan interface{})  //创建一个unbuffered的channel
	go func() {//启动一个goroutine，往s中塞数据
		defer close(s)//退出时关闭chan
		for _, v := range values { //遍历数组
			select {
			case <-done:
				return
			case s <- v: //将数组元素塞到chan中
			}

		}
	}()

	return s
}

/**
	流创建好以后，该咋处理呢？下面我再给你介绍下实现流的方法。
	takeN：只取流中的前 n 个数据；
	takeFn：筛选流中的数据，只保留满足条件的数据；
	takeWhile：只取前面满足条件的数据，一旦不满足条件，就不再取；
	skipN：跳过流中前几个数据；
	skipFn：跳过满足条件的数据；
	skipWhile：跳过前面满足条件的数据，一旦不满足条件，当前这个元素和以后的元素都会输出给 Channel 的 receiver。
 */

// takeN example 别的都差不多
func takeN(done <-chan struct{}, valueStream <-chan interface{}, num int) <-chan interface{}{
	takeStream := make(chan interface{}) //创建输出流
	go func() {
		defer close(takeStream)
		for i := 0; i < num; i++ { //只读取前num个元素
			select {
			case <-done:
				return
			case takeStream <- <-valueStream: //从输入流中读取元素
			}
		}
	}()
	return takeStream
}


/**
	Map-Reduce:不过，我要讲的并不是分布式的 map-reduce，而是单机单进程的 map-reduce 方法。map-reduce 分为两个步骤，第一步是映射（map），处理队列中的数据，第二步是规约（reduce），把列表中的每一个元素按照一定的处理方式处理成结果，放入到结果队列中。就像做汉堡一样，map 就是单独处理每一种食材，reduce 就是从每一份食材中取一部分，做成一个汉堡。
 */

//我们先来看下 map 函数的处理逻辑:
func mapChan(in <-chan interface{}, fn func(interface{}) interface{}) <-chan interface{} {
	out := make(chan interface{}) //创建一个输出chan
	if in == nil { //异常检查
		close(out)
		return out
	}
	go func() { //启动一个goroutine，实现map的主要逻辑
		defer close(out)
		for v := range in { //从输入chan读取数据，执行业务操作，也就是map操作
			out <- fn(v)
		}
	}()

	return out
}

//reduce 函数的处理逻辑如下：
func reduce(in <-chan interface{}, fn func(r,v interface{}) interface{}) interface{} {
	if in == nil { //异常检查
		return nil
	}
	out := <-in //先读取一个元素
	for v := range in{ //实现reduce的主要逻辑
		out = fn(out, v)
	}
	return out
}

//生成一个数据流
func asStreamGenerate(done <-chan struct{}) <-chan interface{}{
	s := make(chan interface{})
	values := []int{1,2,3,4,5}
	go func() {
		defer close(s)
		for _, v := range values{
			select {
			case <-done:
				return
			case s <- v:
			}
		}
	}()

	return s
}

func exampleMapChan()  {
	in := asStreamGenerate(nil)

	//map操作：乘以10
	mapFn := func(v interface{}) interface{}{
		return v.(int) * 10
	}

	//reduce操作：对map的结果进行累加
	reduceFn := func(r, v interface{}) interface{} {
		return r.(int) + v.(int)
	}

	sum := reduce(mapChan(in, mapFn), reduceFn) //返回累加结果
	fmt.Println(sum)
}

//channel 总结 /images/channelTwo.jpg