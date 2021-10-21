package race_test

import (
	"go-learn.com/v1/biz/race"
	"testing"
)

//TestRace 互斥锁的使用
func TestRace(t *testing.T)  {
	race.Race()
}

//TestRaceStruct 结构体内使用互斥锁
func TestRaceStruct(t *testing.T)  {
	race.RaceStruct()
}

//问题
// 1. 目前Mutex的state字段有几个意义，这几个意义分别是由那些字段表示的
// 2. 等待一个Mutex 的 goroutine数最大是多少？能否满足现实的需求