package once

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

/**
 @desc once 简约而不简单的并发原语
 @date 2021-10-26
 @author cf
 */

// Once 可以用来执行且仅仅执行一次动作，常常用于单例对象的初始化场景


// 常用初始化方式： 定义一个package包，使用init()函数进行初始化， 在main()函数中进行初始化

//但是对于延迟初始化，我们会使用下面的方式
/*var connMu sync.Mutex
var conn net.Conn

func getConn() net.Conn {
	connMu.Lock()
	defer connMu.Unlock()

	// 返回创建好的连接
	if conn != nil {
		return conn
	}

	//创建连接
	conn, _ := net.DialTimeout("tcp", "baidu.com:80", 10*time.Second)
	return conn
}

func useConn()  {
	conn := getConn()
	if conn == nil {
		panic("conn is nil")
	}
}

//这种方式虽然实现起来简单，但是有性能问题。一旦连接创建好，每次请求的时候还是得竞争锁才能读取到这个连接，这是比较浪费资源的，因为连接如果创建好是不需要锁保护的
*/

// Once 第一次调用Do方法的时候参数f才会执行，即使后面有N次调用 f参数也不会被执行

func ExampleOnce() {
	var once sync.Once
	f1 := func() {
		fmt.Println("in f1")
	}
	once.Do(f1) //打印出 in f1

	//初始化第二个函数
	f2 := func() {
		fmt.Println("in f2")
	}

	once.Do(f2) //无输出
}

//使用闭包的方式初始化外部资源

//两种错误使用方式
//1、在f中直接或间接执行 once.Do() 会造成死锁问题

func ExampleOnceError()  {
	var once sync.Once
	once.Do(func() {
		once.Do(func() {
			fmt.Println("初始化")
		})
	})
}

//2、未初始化错误 f执行失败
func ExampleOnceInit()  {
	var once sync.Once
	var googleConn net.Conn //到谷歌网站的一个连接
	once.Do(func() {
		googleConn, _= net.Dial("tcp", "")
	})

	googleConn.Write([]byte("GET / HTTP/1.1\r\nHost: google.com\r\n Accept: */*\r\n\r\n"))
	io.Copy(os.Stdout, googleConn)
}