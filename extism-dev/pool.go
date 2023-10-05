package main

import "sync"

type Pool struct {
	wg      sync.WaitGroup
	current int
	max     int
}

func NewPool(max int) *Pool {
	return &Pool{
		wg:      sync.WaitGroup{},
		current: 0,
		max:     max,
	}
}

func RunTask[T any](pool *Pool, f func(T), x T) {
	if pool.max <= 1 {
		f(x)
		return
	}
	if pool.current >= pool.max {
		pool.Wait()
	}
	pool.current += 1
	pool.wg.Add(1)
	go func(x T) {
		defer pool.wg.Done()
		f(x)
	}(x)
}

func (pool *Pool) Wait() {
	pool.wg.Wait()
	pool.current = 0
}
