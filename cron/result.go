package cron

// Result of an event execution
type Result string

const (
	// Success indicates that the event was executed successfully with no error
	Success Result = "success"
	// Failure indicates that the event failed to execute due to an error
	Failure Result = "failure"
)
