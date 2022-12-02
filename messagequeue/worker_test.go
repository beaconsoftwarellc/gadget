package messagequeue

import (
	"sync"
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

func TestWorker_Add(t *testing.T) {
	assert := assert1.New(t)
	wg := &sync.WaitGroup{}
	pool := make(chan *Worker, 1)
	AddWorker(wg, pool)
	workComplete := 0
	var work Work = func() bool {
		workComplete += 1
		return true
	}
	expected := 5
	var w *Worker
	for i := 0; i < expected; i++ {
		w = <-pool
		w.Add(&work)
	}
	w = <-pool
	w.Exit()
	wg.Wait()
	assert.Equal(workComplete, expected)
}

func TestWorker_Multi(t *testing.T) {
	assert := assert1.New(t)
	wg := &sync.WaitGroup{}
	workerCount := 20
	pool := make(chan *Worker, workerCount)
	for i := 0; i < workerCount; i++ {
		AddWorker(wg, pool)
	}
	workComplete := 0
	expected := 0
	var w *Worker
	for i := 0; i < expected; i++ {
		expected += i
		w = <-pool
		var work Work = func() bool {
			workComplete += i
			return true
		}
		w.Add(&work)
	}
	for i := 0; i < workerCount; i++ {
		w = <-pool
		w.Exit()
	}
	wg.Wait()
	assert.Equal(expected, workComplete)
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
