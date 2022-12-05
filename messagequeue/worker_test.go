package messagequeue

import (
	"sync"
	"sync/atomic"
	"testing"

	assert1 "github.com/stretchr/testify/assert"
)

func getWorker(assert *assert1.Assertions, pool chan *Worker) *Worker {
	select {
	case w := <-pool:
		return w
	default:
		assert.Fail("worker should have been added to channel")
	}
	return nil
}

func TestAddWorker(t *testing.T) {
	assert := assert1.New(t)
	wg := &sync.WaitGroup{}
	pool := make(chan *Worker, 1)
	AddWorker(wg, pool)
	getWorker(assert, pool)
	close(pool)
}

type increment struct {
	value  atomic.Int32
	rvalue bool
}

func (a *increment) Work() bool {
	a.value.Add(1)
	return a.rvalue
}

func TestWorker_Add(t *testing.T) {
	assert := assert1.New(t)
	wg := &sync.WaitGroup{}
	pool := make(chan *Worker, 1)
	AddWorker(wg, pool)
	expected := 5
	job := &increment{rvalue: true}
	var w *Worker
	for i := 0; i < expected; i++ {
		w = <-pool
		w.Add(job)
	}
	w = <-pool
	w.Exit()
	wg.Wait()
	assert.Equal(int32(expected), job.value.Load())
}

func TestWorker_Multi(t *testing.T) {
	assert := assert1.New(t)
	wg := &sync.WaitGroup{}
	workerCount := 20
	pool := make(chan *Worker, workerCount)
	for i := 0; i < workerCount; i++ {
		AddWorker(wg, pool)
	}
	job := &increment{rvalue: true}
	expected := 50
	var w *Worker
	for i := 0; i < expected; i++ {
		w = <-pool
		w.Add(job)
	}
	for i := 0; i < workerCount; i++ {
		w = <-pool
		w.Exit()
	}
	wg.Wait()
	assert.Equal(int32(expected), job.value.Load())
}

func TestWorker_Exit(t *testing.T) {
	assert := assert1.New(t)
	wg := &sync.WaitGroup{}
	pool := make(chan *Worker, 2)
	AddWorker(wg, pool)
	w := getWorker(assert, pool)
	w.Exit()
	select {
	case <-pool:
		assert.Fail("pool should be empty")
	default:
		// noop
	}
	// make sure calling exit again does not explode
	w.Exit()
}
