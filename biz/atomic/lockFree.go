package atomic

import (
	"sync/atomic"
	"unsafe"
)

//使用atomic 实现 Lock-Free queue
/**
	这个 lock-free 的实现使用了一个辅助头指针（head），头指针不包含有意义的数据，只是一个辅助的节点，这样的话，出队入队中的节点会更简单。
	入队的时候，通过 CAS 操作将一个元素添加到队尾，并且移动尾指针。
	出队的时候移除一个节点，并通过 CAS 操作移动 head 指针，同时在必要的时候移动尾指针。
 */

//lock-free 的 queue
type LKQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

//通过链表实现，这个数据结构代表链表中的节点
type node struct {
	value interface{}
	next unsafe.Pointer
}

//创建新的链接
func NewLKQueue() *LKQueue {
	n := unsafe.Pointer(&node{})

	return &LKQueue{head: n, tail: n}
}

//入队
func (q *LKQueue) Enqueue(v interface{}) {
	n := &node{value: v}
	for  {
		tail := load(&q.tail)
		next := load(&tail.next)
		if tail == load(&q.tail) { //尾还是尾
			if next == nil { //还没有新数据入队
				if cas(&tail.next, next, n) { //增加到队尾
					cas(&q.tail, tail, n) //入队成功，移动尾巴指针
					return
				}
			}
		}else { //已有新数据加到队列后面，需要移动尾指针
			cas(&q.tail, tail, next)
		}
	}
}

//出队，没有元素则返回nil
func (q *LKQueue) Dequeue() interface{} {
	for  {
		head := load(&q.tail)
		tail := load(&q.tail)
		next := load(&head.next)
		if head == load(&q.head) { //head还是那个head
			if head == tail { //head 和 tail一样
				if next == nil { //空队列
					return nil
				}
				//只是尾指针还没有调整，尝试调整它指向下一个
				cas(&q.tail, tail, next)
			}else {
				//读取出队的数据
				v := next.value
				// 既然要出队了，头指针移动到下一个
				if cas(&q.head, head, next) {
					return v // dequeue is done. return
				}
			}
		}
	}
}


//将 unsafe.Pointer原子加载转换成node
func load(p *unsafe.Pointer) (n *node) {
	return (*node)(atomic.LoadPointer(p))
}

// 封装cas,避免直接将*node转换成unsafe.Pointer
func cas(p *unsafe.Pointer, old, new *node) (ok bool) {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}