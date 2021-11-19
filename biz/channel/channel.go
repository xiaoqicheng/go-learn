package channel

import (
	"fmt"
	"time"
)

/**
	@desc channel 类型学习
	@date 2021-11-18
	@author cf
 */


//channel 的应用场景

/**
	1、数据交流：当作并发的 buffer 或者 queue，解决生产者 - 消费者问题。多个 goroutine 可以并发当作生产者（Producer）和消费者（Consumer）。
	2、数据传递：一个 goroutine 将数据交给另一个 goroutine，相当于把数据的拥有权 (引用) 托付出去。
	3、信号通知：一个 goroutine可以将信号（closing、closed、data ready等）传递给另一个或者另一组goroutine
	4、任务编排: 可以让一组goroutine按照一定的顺序并发或者串行的执行，这就是编排的功能
	5、锁：利用Channel也可以实现互斥锁的机制
 */

// 基本用法
/*
	//channel 的正确语法
	chan string  //可以发送接收string
	chan<- struct{} //只能发送struct{}
	<-chan int 		// 只能从chan接收int

	//channel 中的元素是任意类型，所以也可能是chan类型，判断 <- 和 chan 的规则， 总是尽量和左边的chan结合
	chan<- chan int   	===== chan<- (chan int)
	chan<- <-chan int	===== chan<- (<-chan int)
	<-chan <-chan int	===== <-chan (<-chan int)
	chan (<-chan int)	===== chan (<-chan int)

===============发送数据、接收数据的用法=================================================
	1、发送数据：
		ch <- 200
	2、接收数据 //接收数据时，还可以返回两个值。第一个值是返回的 chan 中的元素，很多人不太熟悉的是第二个值。第二个值是 bool 类型，代表是否成功地从 chan 中读取到一个值，如果第二个参数是 false，chan 已经被 close 而且 chan 中没有缓存的数据，这个时候，第一个值是零值。所以，如果从 chan 读取到一个零值，可能是 sender 真正发送的零值，也可能是 closed 的并且没有缓存元素产生的零值。
		x, _ := ch	 //吧接收的数据复制给变量x
		foo(<-ch) //把接收的一个数据作为参数传递给函数
		<-ch //丢弃接收的一条数据

	3、其他操作
	func example()  {
		var ch = make(chan int, 10)

		//send 和 recv 都可以作为 select 语句的 case clause，如下面的例子：
		for i :=0; i < 10; i++ {
			select {
			case ch <- i:
			case v := <-ch:
				fmt.Println(v)
			}
		}

		//循环
		for v := range ch {
			fmt.Println(v)
		}

		//忽略读取的值，只是清空chan
		for range ch {

		}
	}
*/


//channel 容易犯的错误
//1、使用channel最常见的错误是panic 和 goroutine 泄露
/**
	panic 的状况
	1、close 为 nil  的chan
	2、send 已经close的chan
	3、close已经close的chan
 */

//goroutine 泄露问题 示例
//如果发生超时，process 函数就返回了，这就会导致 unbuffered 的 chan 从来就没有被读取。我们知道，unbuffered chan 必须等 reader 和 writer 都准备好了才能交流，否则就会阻塞。超时导致未读，结果就是子 goroutine 就阻塞在第 7 行永远结束不了，进而导致 goroutine 泄漏。
func process(timeout time.Duration) bool {
	ch := make(chan bool)
	go func() {
		//模拟处理耗时的业务
		time.Sleep((timeout + time.Second))
		ch <- true
		fmt.Println("exit goroutine")
	}()

	select {
	case result := <-ch :
		return result
	case <- time.After(timeout):
		return false
	}
}

/**
	选择方法：
	1、共享资源的并发访问使用传统并发原语；
	2、复杂的任务编排和消息传递使用 Channel；
	3、消息通知机制使用 Channel，除非只想 signal 一个 goroutine，才使用 Cond；
	4、简单等待所有任务的完成用 WaitGroup，也有 Channel 的推崇者用 Channel，都可以；
	5、需要和 Select 语句结合，使用 Channel；
	6、需要和超时配合时，使用 Channel 和 Context。
 */