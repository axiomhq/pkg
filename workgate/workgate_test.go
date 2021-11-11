package workgate

import (
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
		})
	}

	wg.Wait()
	assert.Equal(t, int32(0), numParallel)
}

func TestWorkGateClose(t *testing.T) {
	gate := New(2)

	gate.DoAsync(func() {
		time.Sleep(time.Millisecond * 100)
	})

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
	})

	_, err := gate.TryDo(func() (interface{}, error) {
		return "foo", nil
	})
	assert.Equal(t, ErrGateFull, err)
}
