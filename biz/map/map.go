package _map

import (
	"fmt"
	"hash/fnv"
	"sync"
)

/**
 @desc map 类型 实现线程安全
 @date 2021-11-03
 @author cf
 */

func exampleMap()  {
	var m = make(map[string]int)

	m["a"] = 0
	fmt.Printf("a=%d; b=%d\n", m["a"], m["b"])

	av, aexisted := m["a"]
	bv, bexisted := m["b"]
	fmt.Printf("a=%d, existed:%t; b=%d, existed=%t\n", av, aexisted, bv, bexisted)
}


//map 扩展 可以支持并发读写

type RWMap struct {
	sync.RWMutex //读写锁保护下面的map字段
	m map[int]int
}

//新建一个RWMap
func NewRWMap(n int) *RWMap  {
	return &RWMap{
		m : make(map[int]int, n),
	}
}

func (m *RWMap) Get(k int) (int, bool) { //从map中读取一个值
	m.RLock()
	defer m.RUnlock()

	v, existed := m.m[k]

	return v, existed
}

func (m *RWMap) Set(k, v int)  { //设置一个键值对
	m.Lock()
	defer m.Unlock()
	m.m[k] = v
}

func (m *RWMap) Delete(k int) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, k)
}

func (m *RWMap) Len() int  {
	m.RLock()
	defer m.RUnlock()

	return len(m.m)
}

func (m *RWMap) Each(f func(k, v int) bool)  {
	m.RLock()
	defer m.RUnlock()

	for k, v := range m.m {
		if ! f(k, v) {
			return
		}
	}
}

//分片加锁：更高效的map,  减少锁的粒度 知名的实现方式 orcaman/concurrent-map

var ShardCount = 32

//分成shardCount 个分片的map
type ConcurrentMap []*ConcurrentMapShared

//通过RWMutex保护的线程安全的分片，包含一个map
type ConcurrentMapShared struct {
	items	map[string]interface{}
	sync.RWMutex	//read write mutex,guards access to internal map
}

//创建并发map
func New() ConcurrentMap {
	m := make(ConcurrentMap, ShardCount)
	for i:=0; i < ShardCount; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]interface{})}
	}
	return m
}

// 根据 key 计算分片索引
func (m ConcurrentMap) GetShard(key string) *ConcurrentMapShared {
	h := fnv.New32()
	k, _ := h.Write([]byte(key))

	return m[uint(k)%uint(ShardCount)]
}

func (m ConcurrentMap) Set(key string, value interface{}) {
	// 根据key计算出对应的分片
	shard := m.GetShard(key)
	shard.Lock() //对这个分片加锁，执行业务操作
	shard.items[key] = value
	shard.Unlock()
}

func (m ConcurrentMap) Get(key string) (interface{}, bool) {
	//根据key计算出对应的分片
	shard := m.GetShard(key)
	shard.RLock()
	//获取值
	val, ok := shard.items[key]
	shard.RUnlock()

	return val, ok
}