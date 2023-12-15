package taskmanager

import (
	"github.com/rs/zerolog"
	taskq "github.com/vmihailenco/taskq/v3"
)

// ClientOps allow functional options to be supplied
// that overwrite default client options.
type ClientOps func(c *options)

// defaultClientOptions will return an clientOptions struct with the default settings
//
// Useful for starting with the default and then modifying as needed
func defaultClientOptions() *options {
	// Set the default options
	return &options{
		debug:           false,
		newRelicEnabled: false,
		taskq: &taskqOptions{
			tasks:  make(map[string]*taskq.Task),
			config: DefaultTaskQConfig("taskq"),
		},
	}
}

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
