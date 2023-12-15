/*
Package taskmanager is the task/job management service layer for concurrent and asynchronous tasks with cron scheduling.
*/
package taskmanager

import (
	"context"

	"github.com/BuxOrg/bux/logging"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	taskq "github.com/vmihailenco/taskq/v3"
)

type (

	// TaskManager implements the Tasker interface
	TaskManager struct {
		options *options
	}

	// options holds all the configuration for the client
	options struct {
		cronService     *cronLocal      // Internal cron job client
		debug           bool            // For extra logs and additional debug information
		logger          *zerolog.Logger // Internal logging
		newRelicEnabled bool            // If NewRelic is enabled (parent application)
		taskq           *taskqOptions   // All configuration and options for using TaskQ
	}

	// taskqOptions holds all the configuration for the TaskQ engine
	taskqOptions struct {
		config *taskq.QueueOptions    // Configuration for the TaskQ engine
		queue  taskq.Queue            // Queue for TaskQ
		tasks  map[string]*taskq.Task // Registered tasks
	}
)

// NewTaskManager creates a new client for all TaskManager functionality
//
// If no options are given, it will use the defaultClientOptions()
// ctx may contain a NewRelic txn (or one will be created)
func NewTaskManager(ctx context.Context, opts ...ClientOps) (Tasker, error) {
	// Create a new tm with defaults
	tm := &TaskManager{options: defaultClientOptions()}

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(tm.options)
	}

	// Set logger if not set
	if tm.options.logger == nil {
		tm.options.logger = logging.GetDefaultLogger()
	}

	// Use NewRelic if it's enabled (use existing txn if found on ctx)
	// ctx = tm.options.getTxnCtx(ctx)

	// Load the TaskQ engine
	if err := tm.loadTaskQ(ctx); err != nil {
		return nil, err
	}

	// Create the cron scheduler
	cr := &cronLocal{}
	cr.New()
	cr.Start()
	tm.options.cronService = cr

	// Return the client
	return tm, nil
}

// Close will close client and any open connections
func (tm *TaskManager) Close(ctx context.Context) error {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("close_taskmanager").End()
	}
	if tm != nil && tm.options != nil {

		// Stop the cron scheduler
		if tm.options.cronService != nil {
			tm.options.cronService.Stop()
			tm.options.cronService = nil
		}

		// Close the taskq queue
		if err := tm.options.taskq.queue.Close(); err != nil {
			return err
		}

		// Empty all values and reset
		tm.options.taskq.config = nil
		tm.options.taskq.queue = nil
	}

	return nil
}

// ResetCron will reset the cron scheduler and all loaded tasks
func (tm *TaskManager) ResetCron() {
	tm.options.cronService.New()
	tm.options.cronService.Start()
}

// Debug will set the debug flag
func (tm *TaskManager) Debug(on bool) {
	tm.options.debug = on
}

// IsDebug will return if debugging is enabled
func (tm *TaskManager) IsDebug() bool {
	return tm.options.debug
}

// DebugLog will display verbose logs
func (tm *TaskManager) DebugLog(text string) {
	if tm.IsDebug() {
		tm.options.logger.Info().Msg(text)
	}
}

// IsNewRelicEnabled will return if new relic is enabled
func (tm *TaskManager) IsNewRelicEnabled() bool {
	return tm.options.newRelicEnabled
}

// Tasks will return the list of tasks
func (tm *TaskManager) Tasks() map[string]*taskq.Task {
	return tm.options.taskq.tasks
}

// Factory will return the factory that is set
func (tm *TaskManager) Factory() Factory {
	if tm.options == nil || tm.options.taskq == nil {
		return FactoryEmpty
	}
	if tm.options.taskq.config.Redis != nil {
		return FactoryRedis
	}
	return FactoryMemory
}

// GetTxnCtx will check for an existing transaction
func (tm *TaskManager) GetTxnCtx(ctx context.Context) context.Context {
	if tm.options.newRelicEnabled {
		txn := newrelic.FromContext(ctx)
		if txn != nil {
			ctx = newrelic.NewContext(ctx, txn)
		}
	}
	return ctx
}
