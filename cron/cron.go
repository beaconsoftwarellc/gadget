package cron

import (
	"sync"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/sliceutil"
)

// LoadEvents for scheduling in an instance of [Cron]
type LoadEvents func(qb.LimitOffset) ([]Event, int, error)

// New creates a new instance of [Cron] with the specified parameters
func New(scheduler Scheduler, loadEvents LoadEvents, logger log.Logger) Cron {
	return &cron{
		scheduler:  scheduler,
		loadEvents: loadEvents,
		logger:     logger,
	}
}

// Cron is an interface for scheduling recurring events to be triggered. When Start is called,
// a channel will be returned that will be populated with [Execution] objects for each event that is triggered
// according to its schedule.
type Cron interface {
	// Start the cron scheduler
	Start() (<-chan *Execution, error)
	// Stop the cron scheduler
	Stop()
	// Schedule an event to be executed on its schedule. This should
	// be called when adding events outside of Reload operations.
	// This method is idempotent and can be used to update existing schedules
	// of events as well as add new events.
	Schedule(event Event)
	// Unschedule an event from the scheduler.
	Unschedule(event string)
}

type eventTimer struct {
	Event         Event
	NextExecution time.Time
	Attempts      uint
	Timer         *time.Timer
}

type cron struct {
	scheduler  Scheduler
	events     *sync.Map
	loadEvents LoadEvents
	triggered  chan *Execution
	logger     log.Logger
}

func (c *cron) load() error {
	// keeping this dead simple, we store a timer on each event with the execution of the event in a scope
	// that is in charge of just that event.
	// This makes it so we do not have to track workers, and events go off in parallel.
	// However, we have no control over how many events are happening at a time.
	limitOffset := qb.NewLimitOffset[int]().SetOffset(0).SetLimit(50)
	for event, err := range sliceutil.Flatten(limitOffset, c.loadEvents) {
		if err != nil {
			return err
		}
		c.schedule(event)
	}
	return nil
}

func (c *cron) Start() (<-chan *Execution, error) {
	if c.events != nil {
		return nil, errors.Newf("[GAD.CRN.75] cron already started")
	}
	c.events = new(sync.Map)
	c.triggered = make(chan *Execution, 1000)
	err := c.load()
	if err != nil {
		c.Stop()
		return nil, err
	}
	return c.triggered, nil
}

func (c *cron) Stop() {
	if c.events == nil {
		return
	}
	c.events.Range(func(key, value interface{}) bool {
		timer, ok := value.(*eventTimer)
		if !ok {
			log.Errorf("[GAD.CRN.83] unexpected type %T for key %v", value, key)
		}
		timer.Timer.Stop()
		return true
	})
	// leave the triggered channel open, we don't want to close it and
	// cause a panic by the caller draining it.
}

func (c *cron) Unschedule(eventID string) {
	event, ok := c.loadAndDelete(eventID)
	if !ok {
		c.logger.Warnf("[GAD.CRN.111] event %s not found", eventID)
		return
	}
	if event != nil {
		event.Timer.Stop()
	}
}

func (c *cron) loadAndDelete(eventID string) (*eventTimer, bool) {
	var (
		obj, ok = c.events.LoadAndDelete(eventID)
		et      *eventTimer
	)
	if ok {
		et = obj.(*eventTimer)
	}
	return et, ok
}

func (c *cron) Schedule(event Event) {
	eventTimer, ok := c.loadAndDelete(event.GetID())
	if ok && eventTimer != nil {
		eventTimer.Timer.Stop()
	}
	c.schedule(event)
}

func (c *cron) schedule(event Event) {
	nextExecution := c.scheduler.GetNextExecution(event)
	c.events.Store(event.GetID(), &eventTimer{
		Event:         event,
		NextExecution: nextExecution,
		Timer:         time.AfterFunc(time.Until(nextExecution), func() { c.trigger(event) }),
	})
}

func (c *cron) trigger(event Event) {
	et, ok := c.loadAndDelete(event.GetID())
	if !ok {
		_ = c.logger.Errorf("[GAD.CRN.76] event %s not found, there may be multiple executions",
			event.GetID())
		et = &eventTimer{Event: event, NextExecution: time.Now().UTC()}
	} else {
		et.Timer.Stop()
	}
	execution := &Execution{
		Event: event.GetID(),
		Time:  et.NextExecution.Unix(),
	}
	select {
	case c.triggered <- execution:
	// noop
	default:
		// just log if the channel is full
		c.logger.Errorf("[GAD.CRN.77] triggered channel full: cannot trigger event %s",
			execution.Event)
	}
	c.schedule(event)
}
