package worker_pool

import "errors"

var (
	ErrInvalidPoolSize = errors.New("worker pool size must be greater than zero")
	ErrNilTask         = errors.New("worker pool task is nil")
	ErrPoolClosed      = errors.New("worker pool is closed")
)
