package cron

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	log2 "github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCron_Empty(t *testing.T) {
	var (
		ctrl       = gomock.NewController(t)
		scheduler  = NewScheduler()
		loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
			return nil, 0, nil
		}
		shouldExecute      = func(Event, time.Time) (bool, error) { return true, nil }
		retryAfter         = time.Minute
		maxAttempts   uint = 3
		complete      <-chan *Execution
		err           error
		log           = log2.NewMockLogger(ctrl)
	)
	cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
	complete, err = cron.Start()
	require.NoError(t, err)
	require.NotNil(t, complete)
	cron.Stop()
}

func TestCron_StartTwice(t *testing.T) {
	var (
		ctrl       = gomock.NewController(t)
		scheduler  = NewScheduler()
		loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
			return nil, 0, nil
		}
		shouldExecute      = func(Event, time.Time) (bool, error) { return true, nil }
		retryAfter         = time.Minute
		maxAttempts   uint = 3
		complete      <-chan *Execution
		log           = log2.NewMockLogger(ctrl)
		err           error
	)
	cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
	complete, err = cron.Start()
	require.NoError(t, err)
	require.NotNil(t, complete)

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
		shouldExecute      = func(Event, time.Time) (bool, error) { return true, nil }
		retryAfter         = time.Minute
		maxAttempts   uint = 3
		log                = log2.NewMockLogger(ctrl)
	)
	cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
	cron.Stop()
}

func TestCron_Load(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			ctrl      = gomock.NewController(t)
			scheduler = NewScheduler()

			retryAfter         = time.Minute
			maxAttempts   uint = 3
			complete      <-chan *Execution
			log           = log2.NewMockLogger(ctrl)
			err           error
			now           = time.Now().UTC()
			shouldExecute = func(Event, time.Time) (bool, error) { return true, nil }
			schedule      = mockSchedule{
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		event := NewMockEvent(ctrl)
		event.EXPECT().GetID().Return("foo").AnyTimes()
		event.EXPECT().GetSchedule().Return(schedule).AnyTimes()
		event.EXPECT().Execute().Return(nil)

		loadEvents := func(lo qb.LimitOffset) ([]Event, int, error) {
			if lo.Offset() == 0 {
				return []Event{event}, 1, nil
			}
			return nil, 1, nil
		}

		cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
		complete, err = cron.Start()
		require.NoError(t, err)

		time.Sleep(2 * time.Hour)
		execution := <-complete
		require.Equal(t, event.GetID(), execution.Event)
		require.Equal(t, Success, execution.Result)

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
			retryAfter         = time.Minute
			maxAttempts   uint = 3
			complete      <-chan *Execution
			log           = log2.NewMockLogger(ctrl)
			err           error
			now           = time.Now().UTC()
			shouldExecute = func(Event, time.Time) (bool, error) { return true, nil }
			schedule      = mockSchedule{
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
		complete, err = cron.Start()
		require.NoError(t, err)

		event := NewMockEvent(ctrl)
		event.EXPECT().GetID().Return("foo").AnyTimes()
		event.EXPECT().GetSchedule().Return(schedule).AnyTimes()
		event.EXPECT().Execute().Return(nil)

		cron.Schedule(event)
		time.Sleep(2 * time.Hour)
		execution := <-complete
		require.Equal(t, event.GetID(), execution.Event)
		require.Equal(t, Success, execution.Result)

		event.EXPECT().Execute().Return(nil)
		time.Sleep(24 * time.Hour)
		execution = <-complete
		require.Equal(t, event.GetID(), execution.Event)
		require.Equal(t, Success, execution.Result)
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
			retryAfter         = time.Minute
			maxAttempts   uint = 3
			complete      <-chan *Execution
			log           = log2.NewMockLogger(ctrl)
			err           error
			now           = time.Now().UTC()
			shouldExecute = func(Event, time.Time) (bool, error) { return true, nil }
			schedule      = mockSchedule{
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
		complete, err = cron.Start()
		require.NoError(t, err)

		event := NewMockEvent(ctrl)
		event.EXPECT().GetID().Return("foo").AnyTimes()
		event.EXPECT().GetSchedule().Return(schedule).MaxTimes(2)
		event.EXPECT().Execute().Return(nil)

		cron.Schedule(event)
		time.Sleep(2 * time.Hour)
		execution := <-complete
		require.Equal(t, event.GetID(), execution.Event)
		require.Equal(t, Success, execution.Result)

		// now update the schedule to the 5th of the month
		schedule = mockSchedule{
			hour:       -1,
			minute:     -1,
			dayOfMonth: 5,
			dayOfWeek:  -1,
			month:      -1,
		}
		event.EXPECT().GetSchedule().Return(schedule).AnyTimes()
		cron.Schedule(event)
		// wait 24 hours and we should not have an execution
		time.Sleep(24 * time.Hour)

		event.EXPECT().Execute().Return(nil)
		// wait 5 days and we should have an execution
		time.Sleep(5 * 24 * time.Hour)
		execution = <-complete
		require.Equal(t, event.GetID(), execution.Event)
		require.Equal(t, Success, execution.Result)
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
			retryAfter         = time.Minute
			maxAttempts   uint = 3
			complete      <-chan *Execution
			log           = log2.NewMockLogger(ctrl)
			err           error
			now           = time.Now().UTC()
			shouldExecute = func(Event, time.Time) (bool, error) { return true, nil }
			schedule      = mockSchedule{
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
		complete, err = cron.Start()
		require.NoError(t, err)

		event := NewMockEvent(ctrl)
		event.EXPECT().GetID().Return("foo").AnyTimes()
		event.EXPECT().GetSchedule().Return(schedule).MaxTimes(2)
		event.EXPECT().Execute().Return(nil)

		cron.Schedule(event)
		time.Sleep(2 * time.Hour)
		execution := <-complete
		require.Equal(t, event.GetID(), execution.Event)
		require.Equal(t, Success, execution.Result)

		cron.Unschedule(event.GetID())
		// we should be able to wait a couple of days with no execution
		time.Sleep(2 * 24 * time.Hour)
		cron.Stop()
	})
}

func TestCronRetry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			ctrl       = gomock.NewController(t)
			scheduler  = NewScheduler()
			loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
				return nil, 0, nil
			}
			retryAfter       = time.Hour
			maxAttempts uint = 3
			complete    <-chan *Execution
			log         = log2.NewMockLogger(ctrl)
			err         error
			now         = time.Now().UTC()

			schedule = mockSchedule{
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		doOnce := true
		shouldExecute := func(Event, time.Time) (bool, error) {
			if doOnce {
				doOnce = false
				return false, errors.New(generator.ID("err"))
			}
			return true, nil
		}

		cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
		complete, err = cron.Start()
		require.NoError(t, err)

		event := NewMockEvent(ctrl)
		event.EXPECT().GetID().Return("foo").AnyTimes()
		event.EXPECT().GetSchedule().Return(schedule).AnyTimes()

		log.EXPECT().Errorf("[GAD.CRN.125] error checking for execution of '%s' (retry in %s): %s",
			gomock.Any(), gomock.Any(), gomock.Any())

		cron.Schedule(event)
		// this will park us after the first attempt but before the second
		time.Sleep(90 * time.Minute)
		// now we should have a second attempt that will succeed
		event.EXPECT().Execute().Return(nil)
		time.Sleep(time.Hour)

		execution := <-complete
		require.Equal(t, event.GetID(), execution.Event)
		require.Equal(t, Success, execution.Result)
	})
}

func TestCronRetryToFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var (
			ctrl       = gomock.NewController(t)
			scheduler  = NewScheduler()
			loadEvents = func(qb.LimitOffset) ([]Event, int, error) {
				return nil, 0, nil
			}
			retryAfter       = time.Hour
			maxAttempts uint = 2
			complete    <-chan *Execution
			log         = log2.NewMockLogger(ctrl)
			err         error
			now         = time.Now().UTC()

			schedule = mockSchedule{
				hour:       int32(now.Hour() + 1),
				minute:     -1,
				dayOfMonth: -1,
				dayOfWeek:  -1,
				month:      -1,
			}
		)
		var errorCount uint = 0
		shouldExecute := func(Event, time.Time) (bool, error) {
			if errorCount < maxAttempts {
				errorCount++
				return false, errors.New(generator.ID("err"))
			}
			return true, nil
		}

		cron := New(scheduler, loadEvents, retryAfter, shouldExecute, maxAttempts, log)
		complete, err = cron.Start()
		require.NoError(t, err)
		require.NotNil(t, complete)

		event := NewMockEvent(ctrl)
		event.EXPECT().GetID().Return("foo").AnyTimes()
		event.EXPECT().GetSchedule().Return(schedule).AnyTimes()

		cron.Schedule(event)
		// this will park us after the first attempt but before the second
		log.EXPECT().Errorf("[GAD.CRN.125] error checking for execution of '%s' (retry in %s): %s",
			gomock.Any(), gomock.Any(), gomock.Any())
		time.Sleep(90 * time.Minute)

		// this will park us after the second attempt
		log.EXPECT().Errorf("[GAD.CRN.125] error checking for execution of '%s' (retry in %s): %s",
			gomock.Any(), gomock.Any(), gomock.Any())
		log.EXPECT().Errorf("[GAD.CRN.126] event %s failed after %d attempts", event.GetID(), gomock.Any())
		time.Sleep(90 * time.Minute)

		// now we expect it to have given up and will retry tomorrow
		time.Sleep(90 * time.Minute)

		// now expect an execution
		event.EXPECT().Execute().Return(nil)
		time.Sleep(24 * time.Hour)

		execution := <-complete
		require.Equal(t, event.GetID(), execution.Event)
		require.Equal(t, Success, execution.Result)
	})
}
