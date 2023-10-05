package main

import (
	"sync"

	"github.com/schollz/progressbar/v3"
)

type Pool struct {
	wg          sync.WaitGroup
	current     int
	max         int
	progressBar *progressbar.ProgressBar
}

func NewPool(max int) *Pool {
	return &Pool{
		wg:      sync.WaitGroup{},
		current: 0,
		max:     max,
	}
}

func (pool *Pool) SetProgress(total int, descr ...string) {
	pool.progressBar = progressbar.Default(int64(total), descr...)
}

func RunTask[T any](pool *Pool, f func(T), x T) {
	if pool.progressBar == nil {
		pool.progressBar = progressbar.Default(-1)
	}
	if pool.max <= 1 {
		pool.progressBar.Add(1)
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
		defer pool.progressBar.Add(1)
		f(x)
	}(x)
}

func (pool *Pool) Wait() {
	pool.wg.Wait()
	pool.current = 0
}
