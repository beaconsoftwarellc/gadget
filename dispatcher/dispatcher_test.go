package dispatcher

import (
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/generator"
)

type GenericTask struct {
	execute func() error
}

func (task *GenericTask) Execute() error {
	return task.execute()
}

func TestDispatchTask(t *testing.T) {
	assert := assert.New(t)
	d := NewDispatcher(10, 1, 1)
	d.Run()
	expected := generator.String(20)
	task := newTestTask(expected, false)
	assert.True(d.Dispatch(task))
	actual := <-(task.Task.(TestTask)).comms
	assert.Equal(expected, actual)
	// calling run again should be fine
	d.Run()
	expected = generator.String(20)
	task = newTestTask(expected, false)
	assert.True(d.Dispatch(task))
	actual = <-(task.Task.(TestTask)).comms
	assert.Equal(expected, actual)
	d.Quit(false)
}

func TestNewDispatcherDoesNotFail(t *testing.T) {
	d := NewDispatcher(1, 1, 10)
	d.Run()
	d.Quit(true)
}

func TestDispatchTasks(t *testing.T) {
	assert := assert.New(t)
	d := NewDispatcher(1000, 10, 100)
	d.Run()
	tasks := []TestTask{}
	values := []string{}
	for i := 0; i < 5000; i++ {
		values = append(values, strconv.Itoa(i))
	}
	for _, v := range values {
		task := newTestTask(v, false)
		d.Dispatch(task)
		tasks = append(tasks, (task.Task.(TestTask)))
	}
	actual := []string{}
	for _, task := range tasks {
		actual = append(actual, <-task.comms)
	}
	assert.Equal(values, actual)
	d.Quit(true)
}

func TestMinWorkerSize(t *testing.T) {
	assert := assert.New(t)
	minWorkers := 5
	maxWorkers := minWorkers + 10
	obj := NewDispatcher(1000, minWorkers, maxWorkers)
	d, ok := obj.(*dispatcher)
	assert.True(ok)
	d.Run()
	assert.Equal(minWorkers, len(d.workers))
	d.Quit(false)
}

func TestMaxWorkerSize(t *testing.T) {
	assert := assert.New(t)
	minWorkers := 1
	maxWorkers := 100
	// use 10 buffered doc to slow things down a bit
	obj := NewDispatcher(10, minWorkers, maxWorkers)
	d, ok := obj.(*dispatcher)
	assert.True(ok)
	d.Run()
	scaledUp := false
	for i := 0; i < 100; i++ {
		task := &GenericTask{
			execute: func() error {
				// after a few tasks the pool should have scaled up
				if len(d.workers) > minWorkers {
					scaledUp = true
				}
				assert.True(len(d.workers) <= maxWorkers)
				return nil
			},
		}
		d.Dispatch(task)
	}
	d.Quit(true)
	assert.True(scaledUp)
}

func TestQuitExitsWithoutCompletingQueue(t *testing.T) {
	assert := assert.New(t)
	obj := NewDispatcher(1, 1, 2)
	d, ok := obj.(*dispatcher)
	assert.True(ok)
	d.Run()
	var completedTasks int32
	taskCount := 1000
	for i := 0; i < taskCount; i++ {
		task := &GenericTask{
			execute: func() error {
				atomic.AddInt32(&completedTasks, 1)
				return nil
			},
		}
		d.Dispatch(task)
	}
	d.Quit(false)
	assert.True(atomic.LoadInt32(&completedTasks) < int32(taskCount))
}

func TestAddTaskWhileDrainingStillCompletes(t *testing.T) {
	assert := assert.New(t)
	d := NewDispatcher(100, 1, 20)
	d.Run()
	var completedTasks int32
	var taskCount int32 = 1000
	var i int32
	newTask := func(c *int32) Task {
		return &GenericTask{
			execute: func() error {
				atomic.AddInt32(c, 1)
				return nil
			},
		}
	}
	for ; i < taskCount; i++ {
		d.Dispatch(newTask(&completedTasks))
	}
	go d.Dispatch(newTask(&completedTasks))
	d.Quit(true)
	assert.True(atomic.LoadInt32(&completedTasks) > 900)
}

func TestQuitDrainDrains(t *testing.T) {
	t.SkipNow()
	assert := assert.New(t)
	d := NewDispatcher(100, 1, 30)
	d.Run()
	var completedTasks int32
	var taskCount int32 = 1000
	var i int32
	for ; i < taskCount; i++ {
		task := &GenericTask{
			execute: func() error {
				atomic.AddInt32(&completedTasks, 1)
				return nil
			},
		}
		d.Dispatch(task)
	}
	assert.Equal(taskCount, atomic.LoadInt32(&completedTasks))
}
