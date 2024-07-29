package pool

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

const (
	defaultCapacity = 100
	maxCapacity     = 1000
)

var ErrPoolFreed = errors.New("pool is freed")

type Task func()

type pool struct {
	capacity int            // 协程池的容量
	active   chan struct{}  // 标识活跃worker的chan
	tasks    chan Task      // 任务通道，不设置缓存
	quit     chan struct{}  // 退出通知通道，close无缓冲通道会向所有接受者发消息
	wg       sync.WaitGroup // 等待任务退出
}

// // 创建协程池
func NewPool(cap int) *pool {
	if cap <= 0 {
		cap = defaultCapacity
	}
	if cap >= maxCapacity {
		cap = maxCapacity
	}
	p := &pool{
		capacity: cap,
		active:   make(chan struct{}, cap),
		tasks:    make(chan Task),
		quit:     make(chan struct{}),
	}
	log.Println("--create a new pool--")
	// 启动协程池
	go p.run()

	return p
}

// 启动协程池
func (p *pool) run() {
	var idx int
	for {
		select {
		case p.active <- struct{}{}:
			idx++
			p.newWorker(idx)

		case <-p.quit:
			return
		}
	}
}

// 创建 worker
func (p *pool) newWorker(idx int) {
	// 让 worker 保持工作
	p.wg.Add(1) // 让 quit

	// 多个 worker 不阻塞，开协程
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("worker [%d]: recover panic [%s] and exit \n", idx, err)
				<-p.active
			}
			// 这个 worker 退出后 done -1
			p.wg.Done()
		}()

		log.Printf("new worker: %d\n", idx)

		for {
			select {
			case t := <-p.tasks:
				fmt.Printf("worker [%d]: receive a task\n", idx)
				t()
			case <-p.quit:
				log.Printf("worker [%d] is done\n", idx)
				<-p.active
				return
			}

		}
	}()
}

// 分配任务
func (p *pool) Schedule(t Task) error {
	select {
	case p.tasks <- t:
		return nil
	case <-p.quit:
		return ErrPoolFreed
	}
}

// 释放协程池

func (p *pool) Free() {
	close(p.quit)
	p.wg.Wait()
}
