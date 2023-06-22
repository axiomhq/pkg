package workgate

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

var (
	// ErrGateClosed is returned from WorkGate.{Try,}Do() if the gate has been closed.
	ErrGateClosed = errors.New("gate closed")
	// ErrGateFull is returned from WorkGate.TryDo if the gate is full.
	ErrGateFull = errors.New("gate full")
)

// WorkGate can ensure a maximum of N concurrent tasks are ever ongoing.
// It does _not_ wait for any tasks, use an additional WaitGroup for that.
type WorkGate struct {
	closed uint32
	q      chan struct{}
}

// New creates a new WorkGate.
func New(maxWorkers uint) *WorkGate {
	if maxWorkers == 0 {
		maxWorkers = 1
	}

	q := make(chan struct{}, maxWorkers)

	return &WorkGate{
		q: q,
	}
}

// MaxWorkers returns the maximum number of concurrent tasks.
func (wg *WorkGate) MaxWorkers() int {
	return cap(wg.q)
}

// Enter grabs a token from the WorkGate. If this function returns true the caller is free to do work.
// If it returns false the gate has been closed.
// Caller MUST call wg.Leave() when done (AND Enter() returned true). Low level API, not normally used.
func (wg *WorkGate) Enter() (res bool) {
	defer func() { res = recover() == nil }()
	wg.q <- struct{}{}

	return
}

// Leave must be called when work has been completed after a call to Enter().
// Low level API, not normally used.
func (wg *WorkGate) Leave() {
	<-wg.q
}

// Close prevents further work from being done and the state is permanent. Can be called multiple times.
func (wg *WorkGate) Close() {
	if atomic.SwapUint32(&wg.closed, 1) == 0 {
		close(wg.q)
	}
}

// Do a task on the calling thread. Returns the return value of task.
// Task is silently dropped if gate has been closed.
func (wg *WorkGate) Do(task func() (interface{}, error)) (interface{}, error) {
	if wg.Enter() {
		defer wg.Leave()
		return task()
	}
	return nil, ErrGateClosed
}

// TryDo is like Do, but returns an error if the gate is full.
func (wg *WorkGate) TryDo(task func() (interface{}, error)) (res interface{}, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = ErrGateClosed
		}
	}()

	select {
	case wg.q <- struct{}{}: // same as wg.Enter()
		defer wg.Leave()
		return task()
	default:
		return nil, ErrGateFull
	}
}

// DoAsyncContext is like DoAsync but accepts a context
// if that is cancelled it will stop waiting and the errorHandler fuction will be called with ctx.Err()
// If the gate is closed the errorHandler fuction will be called with ErrGateClosed
// If the task panics the errorHandler fuction will be called with the recovered panic.
// If errorHandler is nil panics won't be recovered.
func (wg *WorkGate) DoAsyncContext(ctx context.Context, task func(), errorHandler func(error)) {
	go func() {
		// Handle wg.q <- struct{}{} panic
		defer func() {
			if r := recover(); r != nil {
				if errorHandler == nil {
					return
				}
				errorHandler(ErrGateClosed)
			}
		}()

		select {
		case <-ctx.Done():
			if errorHandler == nil {
				return
			}
			errorHandler(ctx.Err())
			return
		case wg.q <- struct{}{}: // same as wg.Enter()
			defer func() {
				wg.Leave()
				if errorHandler == nil {
					return
				}
				if r := recover(); r != nil {
					var err error
					if e, ok := r.(error); ok {
						err = fmt.Errorf("panic recovered: %w", e)
					} else {
						err = fmt.Errorf("panic recovered: %v", r)
					}
					errorHandler(err)
				}
			}()
			task()
		}
	}()
}

// DoAsync executes the task in a goroutine.
// Note that it will block if all slots are currently occupied
// If the gate is closed it returns ErrGateClosed
// If the task panics the errorHandler fuction will be called with the recovered panic.
// If errorHandler is nil panics won't be recovered.
func (wg *WorkGate) DoAsync(task func(), errorHandler func(error)) error {
	if wg.Enter() {
		go func() {
			defer func() {
				wg.Leave()
				if errorHandler == nil {
					return
				}
				if r := recover(); r != nil {
					var err error
					if e, ok := r.(error); ok {
						err = fmt.Errorf("panic recovered: %w", e)
					} else {
						err = fmt.Errorf("panic recovered: %v", r)
					}
					errorHandler(err)
				}
			}()
			task()
		}()
		return nil
	}
	return ErrGateClosed
}
