//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination mock$GOFILE
package cron

// Event that can be scheduled for execution at a specific time.
type Event interface {
	// GetID is a unique identifier for this event that can be
	// used to update or remove it from the scheduler.
	GetID() string
	// Execute this event. If this returns an error, the event will
	// be retried according to the [Cron] configured retry schedule.
	Execute() error
	// GetSchedule returns the schedule associated with this event, used to determine its execution timing.
	GetSchedule() Schedule
}
