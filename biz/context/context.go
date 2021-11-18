package context

import (
	"context"
	"fmt"
	"time"
)

/**
	@desc context 上下文信息
	@date 2021-11-17
	@author cf
 */


/**
	常用场景：
	1、上下文信息传递 （request-scoped），比如处理 http 请求、在请求处理链路上传递信息；
	2、控制子 goroutine 的运行；
	3、超时控制的方法调用；
	4、可以取消的方法调用；
 */

var con context.Context
/**
	使用Context的时候，有一些约定俗称的规则
	1、一般函数使用 Context 的时候，会把这个参数放在第一个参数的位置。
	2、从来不把 nil 当做 Context 类型的参数值，可以使用 context.Background() 创建一个空的上下文对象，也不要使用 nil。
	3、Context 只用来临时做函数之间的上下文透传，不能持久化 Context 或者把 Context 长久保存。把 Context 持久化到数据库、本地文件或者全局变量、缓存中都是错误的用法。
	4、key 的类型不应该是字符串类型或者其它内建类型，否则容易在包之间使用 Context 时候产生冲突。使用 WithValue 时，key 的类型应该是自己定义的类型。
	5、常常使用 struct{}作为底层类型定义 key 的类型。对于 exported key 的静态类型，常常是接口或者指针。这样可以尽量减少内存分配。
 */

//创建特殊用途的context的方法: WithValue、WithCancel、WithTimeOut、WithDeadline 包括他们的功能以及实现方式

//WithValue 保存了一个key-value 键值对

//WithCancel 取消长时间的任务

//WithTimeOut 超时时间

//WithDeadline 会返回一个parent副本 并设置了一个不晚于参数d的截止时间，类型为timerCtx


//我们经常使用 Context 来取消一个 goroutine 的运行，这是 Context 最常用的场景之一，Context 也被称为 goroutine 生命周期范围（goroutine-scoped）的 Context，把 Context 传递给 goroutine。但是，goroutine 需要尝试检查 Context 的 Done 是否关闭了：

//example :
func ExampleContext() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer func() {
			fmt.Println("goroutine exit")
		}()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second)
		}
	}

	time.Sleep(time.Second)
	cancel()
	time.Sleep(2 * time.Second)
}