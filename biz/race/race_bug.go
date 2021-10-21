package race

import (
	"fmt"
	"github.com/petermattis/goid"
	"sync"
	"sync/atomic"
	"time"
)

/**
 @desc Mutex 常见的四种错误场景
 @date 2021-10-19
 @author cf
 */

// 1、Lock/Unlock 未成对出现
//example:
func foo()  {
	var mu sync.Mutex
	defer mu.Unlock()
	fmt.Println("hello, world!")
}

// 2、Copy 已使用的Mutex 注意：（Package sync 的同步原语在使用后是不能复制的，是因为 Mutex是一个有状态的对象）
//example:
type counterBug struct {
	sync.Mutex
	countBug int
}

func CopyBug()  {
	var c counterBug
	c.Lock()
	defer c.Unlock()
	c.countBug++
	copyFoo(c)   //复制锁
}

// 这里的counterBug 的参数是通过复制的方式传入的
func copyFoo(c counterBug)  {
	c.Lock()
	defer c.Unlock()
	fmt.Println("in foo")
}


// 3、重入问题  Mutex是不可重入锁！ 一旦重入就会导致报错
func reentryFoo(l sync.Locker)  {
	fmt.Println("in foo")
	l.Lock()
	reentryBar(l)
	l.Unlock()
}

func reentryBar(l sync.Locker)  {
	l.Lock()
	fmt.Println("in bar")
	l.Unlock()
}

func reentryMain()  {
	l := &sync.Mutex{}
	reentryFoo(l)
}

//实现可重入锁！！！！！！！！！！！！！！
//方法1: 通过hacker的方式获取到 goroutine id 记录下获取锁的goroutine id,它可以实现Locker接口
//example：
type RecursiveMutex struct {
	sync.Mutex
	owner int64 //当前持有锁的goroutine id
	recursion int32 //goroutine 重入次数
}

func (m *RecursiveMutex) Lock() {
	gid := goid.Get()
	//如果当前持有锁的goroutine 就是调用这次调用的goroutine，说明就是重入
	if atomic.LoadInt64(&m.owner) == gid {
		m.recursion++
		return
	}

	m.Mutex.Lock()
	// 获取锁的goroutine第一次调用，记录下 gid 并记录调用次数
	atomic.StoreInt64(&m.owner, gid)
	m.recursion = 1
}

func (m *RecursiveMutex) Unlock() {
	gid := goid.Get()
	// 非持有锁的goroutine尝试释放锁，错误的使用
	if atomic.LoadInt64(&m.owner) != gid {
		panic(fmt.Sprintf("wrong thr owner(%d): %d!", m.owner, gid))
	}

	//调用次数减1
	m.recursion--
	//如果goroutine还没有完全释放，直接返回
	if m.recursion != 0 {
		return
	}

	//最后一次调用，释放锁
	atomic.StoreInt64(&m.owner, -1)
	m.Mutex.Unlock()
}

//方案2、调用 lock/unlock时 由goroutine 提供一个token，用来标识它自己，但是这样一来就不满足Locker 接口了
//example:

type TokenRecursiveMutex struct {
	sync.Mutex
	token int64 //设定的token值
	recursion int32 //goroutine 重入次数
}

func (m *TokenRecursiveMutex) Lock(token int64) {
	//如果token 一致，说明就是重入
	if atomic.LoadInt64(&m.token) == token {
		m.recursion++
		return
	}

	//token 不一致，说明不是递归调用
	m.Mutex.Lock()
	// 获取锁后记录token
	atomic.StoreInt64(&m.token, token)
	m.recursion = 1
}

func (m *TokenRecursiveMutex) Unlock(token int64) {
	// 释放其他token持有的锁，错误的使用
	if atomic.LoadInt64(&m.token) != token {
		panic(fmt.Sprintf("wrong thr owner(%d): %d!", m.token, token))
	}

	//调用次数减1
	m.recursion--
	//如果goroutine还没有完全释放，直接返回
	if m.recursion != 0 {
		return
	}

	//最后一次调用，释放锁
	atomic.StoreInt64(&m.token, 0)
	m.Mutex.Unlock()
}

//4、死锁问题
/**
  死锁的原因：
	1、互斥：至少有一个资源是排他性独享的，其他线程必须处于等待状态，直到资源被释放
	2、持有和等待：goroutine 持有一个资源，并且还在请求其它goroutine持有的资源
	3、不可剥夺：资源只能由持有它的goroutine来释放
	4、环路等待：一般来说，存在一组等待进程， P={P1 P2 P3 ...PN}, P1等待P2持有的资源，P2等待P3持有的资源
       以此类推，最后PN等待P1持有的资源，形成了环路等待的死结
 */

//example：
func Deadlock()  {
	//派出所证明
	var psCertificate sync.Mutex
	//物业证明
	var propertyCertificate sync.Mutex

	var wg sync.WaitGroup
	wg.Add(2)

	//派出所处理流程
	go func() {
		defer wg.Done()

		psCertificate.Lock()
		defer psCertificate.Unlock()

		//检查材料
		time.Sleep(time.Second * 5)
		propertyCertificate.Lock()
		propertyCertificate.Unlock()
	}()

	//物业处理流程
	go func() {
		defer wg.Done()

		propertyCertificate.Lock()
		defer propertyCertificate.Unlock()

		//检查材料
		time.Sleep(time.Second * 5)

		psCertificate.Lock()
		psCertificate.Unlock()
	}()

	wg.Wait()
	fmt.Println("success")
}