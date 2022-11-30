package dispatcher

import (
	"errors"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	assert1 "github.com/stretchr/testify/assert"
)

type TestTask struct {
	expected string
	comms    chan string
	doError  bool
}

func (task TestTask) Execute() error {
	task.comms <- task.expected
	var err error
	if task.doError {
		err = errors.New("error")
	}
	return err
}

func newTestTask(expected string, doError bool) *internalTask {
	return newInternalTask(TestTask{comms: make(chan string, 2), expected: expected, doError: doError})
}

func TestAddWorkerForPool(t *testing.T) {
	assert := assert1.New(t)
	pool := make(chan Worker)
	complete := make(chan *internalTask, 50)
	w := NewWorker(pool, complete, log.NewStackLogger())
	actual, ok := w.(*worker)
	assert.True(ok)
	assert.Equal(pool, actual.pool)
}

func TestWorkerRun(t *testing.T) {
	assert := assert1.New(t)
	// we have one message in this test so we need +1
	// so we do not block
	pool := make(chan Worker, 2)
	complete := make(chan *internalTask, 50)
	worker := NewWorker(pool, complete, log.NewStackLogger())
	expected := "foo"
	task := newTestTask(expected, false)
	worker.Run()
	w := <-pool
	w.Exec(task)
	actual := <-(task.Task.(TestTask)).comms
	worker.Quit()
	assert.Equal(actual, expected)
}

func TestWorkerWithErrorMessageContinues(t *testing.T) {
	// we have two messages in this test so we +1 so
	// we do not block
	assert := assert1.New(t)
	pool := make(chan Worker, 3)
	complete := make(chan *internalTask, 50)
	w := NewWorker(pool, complete, log.NewStackLogger())
	expected := "foo"
	errorTask := newTestTask("I throw errors", true)
	task := newTestTask(expected, false)
	w.Run()
	actual, ok := w.(*worker)
	assert.True(ok)
	actual.tasks <- errorTask
	actual.tasks <- task
	<-(errorTask.Task.(TestTask)).comms
	comms := <-(task.Task.(TestTask)).comms
	w.Quit()
	assert.Equal(expected, comms)
}

func TestWorkerExec(t *testing.T) {
	assert := assert1.New(t)
	pool := make(chan Worker, 3)
	complete := make(chan *internalTask, 50)
	w := NewWorker(pool, complete, log.NewStackLogger())
	task := newTestTask("expected", false)
	assert.False(w.Exec(task))

	w.Run()
	w = <-pool
	assert.True(w.Exec(task))

	// calling run again should be fine
	w.Run()
	task = newTestTask("expected", false)
	w = <-pool
	assert.True(w.Exec(task))
	w.Quit()
}
