package cron

import "time"

// Schedule for an Event to be executed on. Negative values indicate that the field is not applicable.
type Schedule interface {
	// GetMinute of execution, [0,59]
	GetMinute() int32
	// GetDayOfWeek of execution [1,7]
	GetDayOfWeek() int32
	// GetDayOfMonth of execution [1,31]
	GetDayOfMonth() int32
	// GetMonth of execution, [1,12]
	GetMonth() int32
	// GetHour of execution [0,23]
	GetHour() int32
}

// Scheduler determines the next execution time based on the provided schedule.
type Scheduler interface {
	// GetNextExecution determines the next execution time based on the provided schedule and returns.
	// All times are UTC.
	GetNextExecution(schedule Schedule) time.Time
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler() Scheduler {
	return &scheduler{timeNow: time.Now}
}

type scheduler struct {
	timeNow func() time.Time
}

// GetNextExecution determines the next execution time based on the provided schedule and returns.
func (s *scheduler) GetNextExecution(schedule Schedule) time.Time {
	var (
		now     = s.timeNow().UTC()
		year    = now.Year()
		month   = now.Month()
		day     = now.Day()
		weekday = now.Weekday()
		hour    = now.Hour()
		minute  = now.Minute()
	)

	if schedule.GetMonth() > 0 {
		// the month has already passed
		if month > time.Month(schedule.GetMonth()) {
			year = year + 1
		}
		month = time.Month(schedule.GetMonth())
		// if the day of the month is specified, check if it has already passed
		// IFF the current month is the month specified
		if schedule.GetDayOfMonth() > 0 {
			day = int(schedule.GetDayOfMonth())
		} else {
			day = 1
		}
		if schedule.GetHour() > 0 {
			hour = int(schedule.GetHour())
		} else {
			hour = 0
		}
		if schedule.GetMinute() > 0 {
			minute = int(schedule.GetMinute())
		} else {
			minute = 0
		}
		next := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
		if now.After(next) || now.Equal(next) {
			next = next.AddDate(1, 0, 0)
		}
		// we are fully configured
		return next
	}

	if schedule.GetDayOfMonth() > 0 {
		day = int(schedule.GetDayOfMonth())
		// day of month and day of week are not compatible, so ignore DayOfWeek
		if schedule.GetHour() > 0 {
			hour = int(schedule.GetHour())
		} else {
			hour = 0
		}
		if schedule.GetMinute() > 0 {
			minute = int(schedule.GetMinute())
		} else {
			minute = 0
		}
		next := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
		if now.After(next) || now.Equal(next) {
			next = next.AddDate(0, 1, 0)
		}
		return next
	}

	if schedule.GetDayOfWeek() >= 0 {
		day += int(time.Weekday(schedule.GetDayOfWeek()) - weekday)
		if schedule.GetHour() > 0 {
			hour = int(schedule.GetHour())
		} else {
			hour = 0
		}
		if schedule.GetMinute() > 0 {
			minute = int(schedule.GetMinute())
		} else {
			minute = 0
		}
		next := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
		if now.After(next) || now.Equal(next) {
			next = next.AddDate(0, 0, 7)
		}
		return next
	}

	if schedule.GetHour() >= 0 {
		hour = int(schedule.GetHour())
		if schedule.GetMinute() >= 0 {
			minute = int(schedule.GetMinute())
		} else {
			minute = 0
		}
		next := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
		if now.After(next) || now.Equal(next) {
			next = next.Add(time.Hour * 24)
		}
		return next
	}

	if schedule.GetMinute() >= 0 {
		minute = int(schedule.GetMinute())
		next := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
		if now.After(next) || now.Equal(next) {
			next = next.Add(time.Hour)
		}
		return next
	}
	return time.Unix(0, 0)
}
