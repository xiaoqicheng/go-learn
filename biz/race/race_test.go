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