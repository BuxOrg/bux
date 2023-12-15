package taskmanager

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	taskq "github.com/vmihailenco/taskq/v3"
)

// ClientOps allow functional options to be supplied
// that overwrite default client options.
type ClientOps func(c *clientOptions)

// defaultClientOptions will return an clientOptions struct with the default settings
//
// Useful for starting with the default and then modifying as needed
func defaultClientOptions() *clientOptions {
	// Set the default options
	return &clientOptions{
		debug:           false,
		newRelicEnabled: false,
		taskq: &taskqOptions{
			tasks:       make(map[string]*taskq.Task),
			factoryType: FactoryMemory,
			config:      DefaultTaskQConfig("taskq"),
		},
	}
}

// GetTxnCtx will check for an existing transaction
func (c *Client) GetTxnCtx(ctx context.Context) context.Context {
	if c.options.newRelicEnabled {
		txn := newrelic.FromContext(ctx)
		if txn != nil {
			ctx = newrelic.NewContext(ctx, txn)
		}
	}
	return ctx
}

// WithNewRelic will enable the NewRelic wrapper
func WithNewRelic() ClientOps {
	return func(c *clientOptions) {
		c.newRelicEnabled = true
	}
}

// WithDebugging will enable debugging mode
func WithDebugging() ClientOps {
	return func(c *clientOptions) {
		c.debug = true
	}
}

// WithTaskQ will override the default TaskQ configuration
func WithTaskQ(config *taskq.QueueOptions, factory Factory) ClientOps {
	return func(c *clientOptions) {
		if config != nil && !factory.IsEmpty() {
			c.taskq.config = config
			c.taskq.factoryType = factory
		}
	}
}

// WithLogger will set the custom logger interface
func WithLogger(customLogger *zerolog.Logger) ClientOps {
	return func(c *clientOptions) {
		if customLogger != nil {
			c.logger = customLogger
		}
	}
}
