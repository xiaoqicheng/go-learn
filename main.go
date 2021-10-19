package main

import (
	"go-learn.com/v1/biz/race"
)

func main() {
	//使用 Mutex 解决并发
	race.Race()
	//struct 使用 Mutex
	race.RaceStruct()
}