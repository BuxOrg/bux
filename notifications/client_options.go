package notifications

import (
	"net/http"
	"time"

	"gorm.io/gorm/logger"
)

const (
	defaultHTTPTimeout = 20 * time.Second
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
		config: &notificationsConfig{
			webhookEndpoint: "",
		},
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

// WithNotifications will set the webhook endpoint
func WithNotifications(webhookEndpoint string) ClientOps {
	return func(c *clientOptions) {
		c.config.webhookEndpoint = webhookEndpoint
	}
}

// WithLogger will set the logger
func WithLogger(log logger.Interface) ClientOps {
	return func(c *clientOptions) {
		c.logger = log
	}
}

// WithDebug will set debugging on notifications
func WithDebug() ClientOps {
	return func(c *clientOptions) {
		c.debug = true
	}
}
