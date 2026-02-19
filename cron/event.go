//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination mock$GOFILE
package cron

// Event that can be scheduled for execution at a specific time.
type Event interface {
	Schedule
	// GetID is a unique identifier for this event that can be
	// used to update or remove it from the scheduler.
	GetID() string
}
