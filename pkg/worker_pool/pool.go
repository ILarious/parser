package worker_pool

import (
	"sync"
)

type WorkerPool struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	isClosed bool
	taskCh   chan func()
	doneCh   chan struct{}
}

func NewWorkerPool(poolSize int) (*WorkerPool, error) {
	if poolSize <= 0 {
		return nil, ErrInvalidPoolSize
	}

	wp := &WorkerPool{
		taskCh: make(chan func(), poolSize),
		doneCh: make(chan struct{}),
	}

	wp.wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		go wp.worker()
	}

	go func() {
		wp.wg.Wait()
		close(wp.doneCh)
	}()

	return wp, nil
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for task := range wp.taskCh {
		task()
	}
}

func (wp *WorkerPool) Submit(task func()) error {
	if task == nil {
		return ErrNilTask
	}

	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.isClosed {
		return ErrPoolClosed
	}

	wp.taskCh <- task
	return nil
}

func (wp *WorkerPool) Close() {
	wp.mu.Lock()
	if wp.isClosed {
		wp.mu.Unlock()
		return
	}
	wp.isClosed = true
	close(wp.taskCh)
	wp.mu.Unlock()
}

func (wp *WorkerPool) Wait() {
	<-wp.doneCh
}
