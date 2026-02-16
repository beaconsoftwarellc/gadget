package cron

// Execution represents an event execution and the result
type Execution struct {
	// Event id that was executed
	Event string
	// Time the event was executed
	Time int64
	// Result of the event execution
	Result Result
	// Error if any occurred
	Error error
}
