package cron

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	log2 "github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mockEvent struct {
	id         string
	minute     int32
	hour       int32
	dayOfMonth int32
	dayOfWeek  int32
	month      int32
}

func (ms *mockEvent) GetID() string        { return ms.id }
func (ms *mockEvent) GetMinute() int32     { return ms.minute }
func (ms *mockEvent) GetDayOfWeek() int32  { return ms.dayOfWeek }
func (ms *mockEvent) GetDayOfMonth() int32 { return ms.dayOfMonth }
func (ms *mockEvent) GetMonth() int32      { return ms.month }
func (ms *mockEvent) GetHour() int32       { return ms.hour }

func TestCron_Empty(t *testing.T) {
	var (
		ctrl       = gomock.NewController(t)
		scheduler  = NewScheduler()
		loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
			return nil, 0, nil
		}
		triggered <-chan *Execution
		err       error
		log       = log2.NewMockLogger(ctrl)
	)
	cron := New(scheduler, loadEvents, log)
	triggered, err = cron.Start()
	require.NoError(t, err)
	require.NotNil(t, triggered)
	cron.Stop()
}

func TestCron_StartTwice(t *testing.T) {
	var (
		ctrl       = gomock.NewController(t)
		scheduler  = NewScheduler()
		loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
			return nil, 0, nil
		}
		triggered <-chan *Execution
		log       = log2.NewMockLogger(ctrl)
		err       error
	)
	cron := New(scheduler, loadEvents, log)
	triggered, err = cron.Start()
	require.NoError(t, err)
	require.NotNil(t, triggered)

	_, err = cron.Start()
	require.Error(t, err)
	cron.Stop()
}

func TestCron_StopNoStart(t *testing.T) {
	var (
		ctrl       = gomock.NewController(t)
		scheduler  = NewScheduler()
		loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
			return nil, 0, nil
		}
		log = log2.NewMockLogger(ctrl)
	)
	cron := New(scheduler, loadEvents, log)
	cron.Stop()
}

func TestCron_Load(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			ctrl      = gomock.NewController(t)
			scheduler = NewScheduler()

			triggered <-chan *Execution
			log       = log2.NewMockLogger(ctrl)
			err       error
			now       = time.Now().UTC()
			event     = &mockEvent{
				id:         generator.ID("eve"),
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)

		loadEvents := func(lo qb.LimitOffset) ([]Event, int, error) {
			if lo.Offset() == 0 {
				return []Event{event}, 1, nil
			}
			return nil, 1, nil
		}

		cron := New(scheduler, loadEvents, log)
		triggered, err = cron.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Hour)
		execution := <-triggered
		require.Equal(t, event.GetID(), execution.Event)
		cron.Stop()
	})
}

func TestCron_Schedule_New(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			ctrl       = gomock.NewController(t)
			scheduler  = NewScheduler()
			loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
				return nil, 0, nil
			}
			triggered <-chan *Execution
			log       = log2.NewMockLogger(ctrl)
			err       error
			now       = time.Now().UTC()
			event     = &mockEvent{
				id:         generator.ID("eve"),
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		cron := New(scheduler, loadEvents, log)
		triggered, err = cron.Start()
		require.NoError(t, err)

		cron.Schedule(event)
		time.Sleep(2 * time.Hour)
		execution := <-triggered
		require.Equal(t, event.GetID(), execution.Event)
		require.Empty(t, triggered)
		time.Sleep(24 * time.Hour)
		execution = <-triggered
		require.Equal(t, event.GetID(), execution.Event)
		cron.Stop()
	})
}

func TestCron_Schedule_Update(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			ctrl       = gomock.NewController(t)
			scheduler  = NewScheduler()
			loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
				return nil, 0, nil
			}
			triggered <-chan *Execution
			log       = log2.NewMockLogger(ctrl)
			err       error
			now       = time.Now().UTC()
			event     = &mockEvent{
				id:         generator.ID("eve"),
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		cron := New(scheduler, loadEvents, log)
		triggered, err = cron.Start()
		require.NoError(t, err)

		cron.Schedule(event)
		time.Sleep(2 * time.Hour)
		execution := <-triggered
		require.Equal(t, event.GetID(), execution.Event)

		// now update the schedule to the 5th of the month
		event.hour = -1
		event.minute = -1
		event.dayOfMonth = 5
		event.dayOfWeek = -1
		event.month = -1
		cron.Schedule(event)
		// wait 24 hours and we should not have an execution
		time.Sleep(24 * time.Hour)
		require.Empty(t, triggered)

		// wait 5 days and we should have an execution
		time.Sleep(5 * 24 * time.Hour)
		execution = <-triggered
		require.Equal(t, event.GetID(), execution.Event)
		cron.Stop()
	})
}

func TestCron_Unschedule(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			ctrl       = gomock.NewController(t)
			scheduler  = NewScheduler()
			loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
				return nil, 0, nil
			}
			triggered <-chan *Execution
			log       = log2.NewMockLogger(ctrl)
			err       error
			now       = time.Now().UTC()
			event     = &mockEvent{
				id:         generator.ID("eve"),
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		cron := New(scheduler, loadEvents, log)
		triggered, err = cron.Start()
		require.NoError(t, err)

		cron.Schedule(event)
		time.Sleep(2 * time.Hour)
		execution := <-triggered
		require.Equal(t, event.GetID(), execution.Event)

		cron.Unschedule(event.GetID())
		// we should be able to wait a couple of days with no execution
		time.Sleep(2 * 24 * time.Hour)
		require.Empty(t, triggered)
		cron.Stop()
	})
}
