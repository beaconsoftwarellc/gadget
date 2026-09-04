package timeutil

import "time"

// Future provides a time.Time duration into the future from now
func Future(duration time.Duration) time.Time {
	return time.Now().UTC().Add(duration)
}

// Past provides a time.Time duration into the past from now
func Past(duration time.Duration) time.Time {
	return time.Now().UTC().Add(-1 * duration)
}

// Today returns the current date rounded down to midnight UTC
func Today() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour)
}
