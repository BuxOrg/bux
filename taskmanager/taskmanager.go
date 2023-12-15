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

	// Client is the taskmanager client (configuration)
	Client struct {
		options *clientOptions
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
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

// NewClient creates a new client for all TaskManager functionality
//
// If no options are given, it will use the defaultClientOptions()
// ctx may contain a NewRelic txn (or one will be created)
func NewClient(_ context.Context, opts ...ClientOps) (TaskManagerInterface, error) {
	// Create a new client with defaults
	client := &Client{options: defaultClientOptions()}

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(client.options)
	}

	// Set logger if not set
	if client.options.logger == nil {
		client.options.logger = logging.GetDefaultLogger()
	}

	// Use NewRelic if it's enabled (use existing txn if found on ctx)
	// ctx = client.options.getTxnCtx(ctx)

	// Load the TaskQ engine
	if err := client.loadTaskQ(); err != nil {
		return nil, err
	}

	// Create the cron scheduler
	cr := &cronLocal{}
	cr.New()
	cr.Start()
	client.options.cronService = cr

	// Return the client
	return client, nil
}

// Close will close client and any open connections
func (c *Client) Close(ctx context.Context) error {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("close_taskmanager").End()
	}
	if c != nil && c.options != nil {

		// Stop the cron scheduler
		if c.options.cronService != nil {
			c.options.cronService.Stop()
			c.options.cronService = nil
		}

		// Close the taskq queue
		if err := c.options.taskq.queue.Close(); err != nil {
			return err
		}

		// Empty all values and reset
		c.options.taskq.config = nil
		c.options.taskq.queue = nil
	}

	return nil
}

// ResetCron will reset the cron scheduler and all loaded tasks
func (c *Client) ResetCron() {
	c.options.cronService.New()
	c.options.cronService.Start()
}

// Debug will set the debug flag
func (c *Client) Debug(on bool) {
	c.options.debug = on
}

// IsDebug will return if debugging is enabled
func (c *Client) IsDebug() bool {
	return c.options.debug
}

// DebugLog will display verbose logs
func (c *Client) DebugLog(text string) {
	if c.IsDebug() {
		c.options.logger.Info().Msg(text)
	}
}

// IsNewRelicEnabled will return if new relic is enabled
func (c *Client) IsNewRelicEnabled() bool {
	return c.options.newRelicEnabled
}

// Tasks will return the list of tasks
func (c *Client) Tasks() map[string]*taskq.Task {
	return c.options.taskq.tasks
}

// Factory will return the factory that is set
func (c *Client) Factory() Factory {
	if c.options == nil || c.options.taskq == nil {
		return FactoryEmpty
	}
	if c.options.taskq.config.Redis != nil {
		return FactoryRedis
	}
	return FactoryMemory
}
