package dateutil

import (
	"time"
)

// IncrementMonth increments a date by the specified number of months safely handling the end-of-month edge case
func IncrementMonth(t time.Time, months int) time.Time {
	nextMonth := t.AddDate(0, months, 0)
	endOfMonth := time.Date(t.Year(),
		t.Month()+time.Month(months+1),
		0,
		t.Hour(),
		t.Minute(),
		t.Second(),
		0, t.Location())
	if nextMonth.After(endOfMonth) {
		return endOfMonth
	}
	return nextMonth
}

// IncrementYear increments a date by the specified number of years safely handling the end-of-month edge case
func IncrementYear(t time.Time, years int) time.Time {
	return IncrementMonth(t, years*12)
}
