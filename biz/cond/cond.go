package cond

/**
 @desc cond原语
 @date 2021-10-25
 @author cf
 */


//Go 标准库提供 Cond 原语的目的是，为等待/通知场景下的并发问题提供支持。Cond 通常应用于等待某个条件的一组 goroutine，等条件变为 true 的时候，其中一个 goroutine 或者所有的 goroutine 都会被唤醒执行。

