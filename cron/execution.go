package cron

// Execution represents an event execution and the result
type Execution struct {
	// Event id that was executed
	Event string
	// Time the event was executed
	Time int64
}
