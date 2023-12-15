package taskmanager

import (
	"github.com/rs/zerolog"
	taskq "github.com/vmihailenco/taskq/v3"
)

// ClientOps allow functional options to be supplied
// that overwrite default client options.
type ClientOps func(c *options)

// WithNewRelic will enable the NewRelic wrapper
func WithNewRelic() ClientOps {
	return func(c *options) {
		c.newRelicEnabled = true
	}
}

// WithDebugging will enable debugging mode
func WithDebugging() ClientOps {
	return func(c *options) {
		c.debug = true
	}
}

// WithTaskqConfig will set the taskq custom config
func WithTaskqConfig(config *taskq.QueueOptions) ClientOps {
	return func(c *options) {
		if config != nil {
			c.taskq.config = config
		}
	}
}

// WithLogger will set the custom logger interface
func WithLogger(customLogger *zerolog.Logger) ClientOps {
	return func(c *options) {
		if customLogger != nil {
			c.logger = customLogger
		}
	}
}
