package workgate

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkGateDo(t *testing.T) {
	totalTestDuration := time.Second
	numWorkers := int32(1000)
	maxWorkers := int32(7)
	gate := New(uint(maxWorkers))
	assert.EqualValues(t, maxWorkers, gate.MaxWorkers())

	var numParallel int32
	wg := sync.WaitGroup{}

	// Sleep time if we run sequentially == totalTestDuration / numWorkers
	// But since we have maxWorkers in parallel we must multiply byt that number
	workerSleepTime := int64(maxWorkers) * (int64(totalTestDuration) / int64(numWorkers))

	// Spawn parallel workers galore ensuring we never have more than maxWorkers concurrently
	for i := int32(0); i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			endState, err := gate.Do(func() (interface{}, error) {
				assert.LessOrEqual(t, atomic.AddInt32(&numParallel, 1), maxWorkers)

				time.Sleep(time.Duration(workerSleepTime))

				state := atomic.AddInt32(&numParallel, -1)
				assert.LessOrEqual(t, state, maxWorkers)
				assert.GreaterOrEqual(t, state, int32(0))
				wg.Done()
				return state, nil
			})

			assert.LessOrEqual(t, endState, maxWorkers)
			assert.GreaterOrEqual(t, endState, int32(0))
			assert.Nil(t, err)
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(0), numParallel)
}

func TestWorkGateDoAsync(t *testing.T) {
	totalTestDuration := time.Second
	numWorkers := int32(1000)
	maxWorkers := int32(7)
	gate := New(uint(maxWorkers))
	assert.EqualValues(t, maxWorkers, gate.MaxWorkers())

	var numParallel int32
	wg := sync.WaitGroup{}

	// Sleep time if we run sequentially == totalTestDuration / numWorkers
	// But since we have maxWorkers in parallel we must multiply byt that number
	workerSleepTime := int64(maxWorkers) * (int64(totalTestDuration) / int64(numWorkers))

	// Spawn parallel workers galore ensuring we never have more than maxWorkers concurrently
	for i := int32(0); i < numWorkers; i++ {
		wg.Add(1)
		gate.DoAsync(func() {
			assert.LessOrEqual(t, atomic.AddInt32(&numParallel, 1), maxWorkers)

			time.Sleep(time.Duration(workerSleepTime))

			state := atomic.AddInt32(&numParallel, -1)
			assert.LessOrEqual(t, state, maxWorkers)
			assert.GreaterOrEqual(t, state, int32(0))
			wg.Done()
		}, nil)
	}

	wg.Wait()
	assert.Equal(t, int32(0), numParallel)
}

func TestWorkGateClose(t *testing.T) {
	gate := New(2)

	gate.DoAsync(func() {
		time.Sleep(time.Millisecond * 100)
	}, nil)

	gate.Close()

	_, postCloseErr := gate.Do(func() (interface{}, error) { return nil, nil })
	assert.EqualError(t, postCloseErr, ErrGateClosed.Error())

	// make sure the async task finishes and there's no panic
	time.Sleep(time.Millisecond * 200)

	// close one more time (shouldn't panic)
	gate.Close()
	assert.EqualValues(t, 1, gate.closed)
}

func TestWorkGateFull(t *testing.T) {
	gate := New(1)
	defer gate.Close()

	wait := make(chan struct{})
	defer close(wait)
	gate.DoAsync(func() {
		<-wait
	}, nil)

	_, err := gate.TryDo(func() (interface{}, error) {
		return "foo", nil
	})
	assert.Equal(t, ErrGateFull, err)
}

func TestResourceGateDoAsyncRecover(t *testing.T) {
	gate := New(1000)

	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		i2 := i
		gate.DoAsync(
			func() {
				if i2%2 == 1 {
					panic("its odd")
				}
				errs <- nil
			},
			func(err error) {
				errs <- err
			},
		)
	}
	var panics, succeses int
	for i := 0; i < 10; i++ {
		err := <-errs
		if err != nil {
			assert.Equal(t, err, errors.New("panic recovered: its odd"))
			panics++
		} else {
			succeses++
		}
	}
	assert.EqualValues(t, 5, panics)
	assert.EqualValues(t, 5, succeses)
}

func TestResourceGateDoAsyncContextRecover(t *testing.T) {
	gate := New(1000)

	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		i2 := i
		gate.DoAsyncContext(
			context.TODO(),
			func() {
				if i2%2 == 1 {
					panic("its odd")
				}
				errs <- nil
			},
			func(err error) {
				errs <- err
			},
		)
	}
	var panics, succeses int
	for i := 0; i < 10; i++ {
		err := <-errs
		if err != nil {
			assert.Equal(t, err, errors.New("panic recovered: its odd"))
			panics++
		} else {
			succeses++
		}
	}
	assert.EqualValues(t, 5, panics)
	assert.EqualValues(t, 5, succeses)
}

func TestResourceGateDoAsyncRecoverWithError(t *testing.T) {
	gate := New(1000)
	wg := sync.WaitGroup{}
	wg.Add(1)
	gate.DoAsync(
		func() { panic(errors.New("error")) },
		func(err error) {
			assert.Equal(t, errors.Unwrap(err), errors.New("error"))
			wg.Done()
		},
	)
	wg.Wait()
}

func TestResourceGateDoAsyncContextRecoverWithError(t *testing.T) {
	gate := New(1000)
	wg := sync.WaitGroup{}
	wg.Add(1)
	gate.DoAsyncContext(
		context.TODO(),
		func() { panic(errors.New("error")) },
		func(err error) {
			assert.Equal(t, errors.Unwrap(err), errors.New("error"))
			wg.Done()
		},
	)
	wg.Wait()
}

func TestResourceGateDoAsyncContextCanceledContext(t *testing.T) {
	gate := New(50)
	ctx, cancel := context.WithCancel(context.Background())
	var c atomic.Int32
	var e atomic.Int32
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		gate.DoAsyncContext(ctx,
			func() {
				time.Sleep(100 * time.Millisecond)
				c.Add(1)
				wg.Done()
			},
			func(err error) {
				assert.Equal(t, err, errors.New("context canceled"))
				c.Add(1)
				e.Add(1)
				wg.Done()
			},
		)
	}
	cancel()

	wg.Wait()
	assert.EqualValues(t, 100, c.Load())
	assert.GreaterOrEqual(t, e.Load(), int32(1))
}

func TestResourceGateDoAsyncGateClosed(t *testing.T) {
	gate := New(1000)
	gate.Close()
	err := gate.DoAsync(
		func() {},
		nil,
	)
	assert.Equal(t, err, ErrGateClosed)
}

func TestResourceGateDoAsyncContextGateClosed(t *testing.T) {
	gate := New(1000)
	gate.Close()
	wg := sync.WaitGroup{}
	wg.Add(1)
	gate.DoAsyncContext(
		context.TODO(),
		func() { panic(errors.New("error")) },
		func(err error) {
			assert.Equal(t, err, ErrGateClosed)
			wg.Done()
		},
	)
	wg.Wait()
}
