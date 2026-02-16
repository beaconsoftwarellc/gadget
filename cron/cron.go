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

// ShouldExecute is a check executed before an event is executed with the
// time that the execution was for based upon the schedule. This is used
// to prevent duplicate executions by Crons running in a parallel fashion.
type ShouldExecute func(Event, time.Time) (bool, error)

// New creates a new instance of [Cron] with the specified parameters
func New(scheduler Scheduler, loadEvents LoadEvents,
	retryAfter time.Duration, shouldExecute ShouldExecute,
	maxAttempts uint, logger log.Logger) Cron {
	return &cron{
		scheduler:     scheduler,
		loadEvents:    loadEvents,
		retryAfter:    retryAfter,
		maxAttempts:   maxAttempts,
		logger:        logger,
		shouldExecute: shouldExecute,
	}
}

// Cron is an interface for scheduling events to be executed at a specific time.
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
	scheduler     Scheduler
	events        *sync.Map
	loadEvents    LoadEvents
	complete      chan *Execution
	shouldExecute ShouldExecute
	retryAfter    time.Duration
	maxAttempts   uint
	logger        log.Logger
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
		c.schedule(event, 0, nil)
	}
	return nil
}

func (c *cron) Start() (<-chan *Execution, error) {
	if c.events != nil {
		return nil, errors.Newf("[GAD.CRN.75] cron already started")
	}
	c.events = new(sync.Map)
	c.complete = make(chan *Execution, 1000)
	err := c.load()
	if err != nil {
		c.Stop()
		return nil, err
	}
	return c.complete, nil
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
	// leave the complete channel open, we don't want to close it and
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
	c.schedule(event, 0, nil)
}

func (c *cron) schedule(event Event, attempt uint, override *time.Time) {
	nextExecution := c.scheduler.GetNextExecution(event.GetSchedule())
	if override != nil && !override.IsZero() {
		nextExecution = *override
	}
	c.events.Store(event.GetID(), &eventTimer{
		Event:    event,
		Attempts: attempt,
		Timer:    time.AfterFunc(time.Until(nextExecution), func() { c.execute(event) }),
	})
}

// Execute the passed [Execution] and return the next [Execution] for the Event
func (c *cron) execute(event Event) {
	et, ok := c.loadAndDelete(event.GetID())
	if !ok {
		_ = c.logger.Errorf("[GAD.CRN.76] event %s not found, there may be multiple executions",
			event.GetID())
		et = &eventTimer{Event: event, NextExecution: time.Now().UTC()}
	} else {
		et.Timer.Stop()
	}
	if ok, err := c.shouldExecute(event, et.NextExecution); err != nil {
		_ = c.logger.Errorf("[GAD.CRN.125] error checking for execution of '%s' (retry in %s): %s",
			event.GetID(), c.retryAfter, err)
		et.Attempts++
		if et.Attempts >= c.maxAttempts {
			c.logger.Errorf("[GAD.CRN.126] event %s failed after %d attempts", event.GetID(), et.Attempts)
		} else {
			override := time.Now().Add(c.retryAfter)
			c.schedule(event, et.Attempts, &override)
			return
		}
	} else if ok {
		execution := &Execution{
			Event:  event.GetID(),
			Time:   time.Now().Unix(),
			Error:  event.Execute(),
			Result: Success,
		}
		if execution.Error != nil {
			execution.Result = Failure
		}
		select {
		case c.complete <- execution:
		// noop
		default:
			// just log if the channel is full
			c.logger.Errorf("[GAD.CRN.77] execution channel full: execution %s result %s: %s",
				execution.Event, execution.Result, execution.Error)
		}
	}
	c.schedule(event, 0, nil)
}
