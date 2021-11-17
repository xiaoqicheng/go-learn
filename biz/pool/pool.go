package pool

/**
 @desc pool 性能大杀器
 @date 2021-11-09
 @author cf
 */

// sync.Pool 结构 可以创建池化对象，减少创建链接所消耗的时间

//特点： 1. sync.Pool 本身就是线程安全的，多个 goroutine 可以并发地调用它的方法存取对象；
//		2. sync.Pool 不可在使用之后再复制使用。

/**
 // example: 这段代码 只是一个简单示例，可能存在内存泄露问题，不要应用在实际项目中
	var buffers = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	func GetBuffer() *bytes.Buffer {
		return buffers.Get().(*bytes.Buffer)
	}

	func PutBuffer(buf *bytes.Buffer)  {
		buf.Reset()
		buffers.Put(buf)
	}

 */

/*
 sync.Pool 的坑
	1、内存泄露  使用sync.Pool 回收buffer的时候，一定要检查回收的对象大小，如果buffer
	2、内存浪费  可以查看标准库中 net/http/server.go 中的代码
 */

//第三方库的介绍
//1、bytebufferpool 2、oxtoacart/bpool


//TCP连接池  fatih开发的 fatih/pool 它的套路为

/**
	//工厂模式，提供创建连接的工厂方法
	factory := func() (net.Conn, error) { return net.Dial("tcp", "127.0.0.1:40000")}

	//创建一个tcp池，提供初始容量和最大容量以及工厂方法
	p, err := pool.NewChannelPool(5,30,factory)

	//获取一个连接
	conn, err := p.Get()

	// close并不会真正关闭这个连接，而是把它放回池子，所以你不必显式地Put这个对象到池子中
	conn.Close()

	//通过调用MarkUnusable, Close的时候就会真正关闭底层的tcp连接了
	if pc, ok := conn.(*pool.PoolConn); ok {
		pc.MarkUnusable()
		pc.Close()
	}

	//关闭池子就会关闭=池子中的所有的tcp连接
	p.Close()
	// 当前池子中的连接的数量
	current := p.Len()
 */


//Pool分析 images/pool.jpg