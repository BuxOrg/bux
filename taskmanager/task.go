package taskmanager

import (
	"time"
)

// Task is the options for a new task (mimics TaskQ)
type Task struct {
	Name string // Task name.

	// Function called to process a message.
	// There are three permitted types of signature:
	// 1. A zero-argument function
	// 2. A function whose arguments are assignable in type from those which are passed in the message
	// 3. A function which takes a single `*Message` argument
	// The handler function may also optionally take a Context as a first argument and may optionally return an error.
	// If the handler takes a Context, when it is invoked it will be passed the same Context as that which was passed to
	// `StartConsumer`. If the handler returns a non-nil error the message processing will fail and will be retried/.
	Handler interface{}
	// Function called to process failed message after the specified number of retries have all failed.
	// The FallbackHandler accepts the same types of function as the Handler.
	FallbackHandler interface{}

	// Optional function used by Consumer with defer statement to recover from panics.
	DeferFunc func()

	// Number of tries/releases after which the message fails permanently and is deleted. Default is 64 retries.
	RetryLimit int

	// Minimum backoff time between retries. Default is 30 seconds.
	MinBackoff time.Duration

	// Maximum backoff time between retries. Default is 30 minutes.
	MaxBackoff time.Duration
}

// TaskOptions are used for running a task
type TaskOptions struct {
	Arguments      []interface{} `json:"arguments"`        // Arguments for the task
	Delay          time.Duration `json:"delay"`            // Run after X delay
	OnceInPeriod   time.Duration `json:"once_in_period"`   // Run once in X period
	RunEveryPeriod time.Duration `json:"run_every_period"` // Cron job!
	TaskName       string        `json:"task_name"`        // Name of the task
}

/*
// todo: add this functionality to the task options
OnceInPeriod(period time.Duration, args ...interface{})
OnceWithDelay(delay time.Duration)
OnceWithSchedule(tm time.Time)
*/
