package race

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

/**
  Mutex 扩展 尝试获取排它锁
 */

const (
	mutexLocked = 1 << iota // mutex is locked
	mutexWoken //唤醒锁
	mutexStarving //锁饥饿标识位置
	mutexWaiterShift = iota
)

type Mutex struct {
	sync.Mutex
}

func (m *Mutex) tryLock() bool {
	//如果能成功抢到锁
	if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), 0, mutexLocked){
		return true
	}

	//如果处于唤醒、加锁、或者饥饿状态，这次请求就不参与竞争，直接返回false
	old := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))
	if old & (mutexLocked | mutexWoken | mutexStarving) != 0 {
		return false
	}

	// 尝试在竞争的状态下获取锁
	newMutex := old | mutexLocked
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), old, newMutex)
}

//测试
func Try()  {
	var mu Mutex
	go func() {
		mu.Lock()
		time.Sleep(time.Second * time.Duration(rand.Intn(2)))
		mu.Unlock()
	}()

	time.Sleep(time.Second)

	ok := mu.tryLock() //尝试获取锁
	if ok {
		fmt.Println("got the lock")

		//do something

		mu.Unlock()
		return
	}

	//没有获取到
	fmt.Println("cannot get the lock")
}

//获取state这个字段并进行解析
func (m *Mutex) Count() int {
	//获取state字段的值
	v := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.Mutex)))

	v = v >> mutexWaiterShift //得到等待者的数值
	v = v >> v + (v & mutexLocked) //在加上锁持有者的数量 0 || 1

	return int(v)
}



// Mutex 实现线程安全
//example: 队列

type SliceQueue struct {
	data []interface{}
	mu sync.Mutex
}

//创建一个数组队列
func NewSliceQueue(n int) (q *SliceQueue) {
	return &SliceQueue{data: make([]interface{}, 0, n)}
}

func (q *SliceQueue) Enqueue(v interface{}) {
	q.mu.Lock()
	q.data = append(q.data, v)
	q.mu.Unlock()
}

func (q *SliceQueue) Dequeue() interface{} {
	q.mu.Lock()
	if len(q.data) == 0 {
		q.mu.Unlock()
		return nil
	}

	v := q.data[0]
	q.data = q.data[1:]

	q.mu.Unlock()
	return v
}