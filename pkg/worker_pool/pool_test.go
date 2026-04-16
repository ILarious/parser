package worker_pool

import (
	"sync"
	"testing"
)

func TestWorkerPoolSubmitAndWait(t *testing.T) {
	pool, err := NewWorkerPool(2)
	if err != nil {
		t.Fatalf("NewWorkerPool() error = %v", err)
	}

	var (
		mu   sync.Mutex
		got  []int
		done sync.WaitGroup
	)
	done.Add(3)

	for i := 1; i <= 3; i++ {
		i := i
		if err := pool.Submit(func() {
			defer done.Done()
			mu.Lock()
			got = append(got, i)
			mu.Unlock()
		}); err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	done.Wait()
	pool.Close()
	pool.Wait()

	if len(got) != 3 {
		t.Fatalf("expected 3 tasks executed, got %d", len(got))
	}
}

func TestWorkerPoolRejectsSubmitAfterClose(t *testing.T) {
	pool, err := NewWorkerPool(1)
	if err != nil {
		t.Fatalf("NewWorkerPool() error = %v", err)
	}

	pool.Close()
	pool.Wait()

	if err := pool.Submit(func() {}); err != ErrPoolClosed {
		t.Fatalf("Submit() error = %v, want %v", err, ErrPoolClosed)
	}
}
