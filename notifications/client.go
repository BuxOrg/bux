package notifications

import (
	zLogger "github.com/mrz1836/go-logger"
)

// EventType event types thrown in Bux
type EventType string

const (
	// EventTypeCreate when a new model is created
	EventTypeCreate EventType = "create"

	// EventTypeUpdate when a new model is updated
	EventTypeUpdate EventType = "update"

	// EventTypeDelete when a new model is deleted
	EventTypeDelete EventType = "delete"

	// EventTypeBroadcast when a transaction is broadcasted (sync tx)
	EventTypeBroadcast EventType = "broadcast"
)

type (

	// Client is the client (configuration)
	Client struct {
		options *clientOptions
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
		config     *notificationsConfig        // Configuration for broadcasting and other chain-state actions
		debug      bool                        // Debugging mode
		httpClient HTTPInterface               // Custom HTTP client
		logger     zLogger.GormLoggerInterface // Custom logger interface
	}

	// syncConfig holds all the configuration about the different notifications
	notificationsConfig struct {
		webhookEndpoint string // Webhook URL for basic notifications
	}
)

// NewClient creates a new client for notifications
func NewClient(opts ...ClientOps) (ClientInterface, error) {

	// Create a new client with defaults
	client := &Client{
		options: defaultClientOptions(),
	}

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(client.options)
	}

	// Set logger if not set
	if client.options.logger == nil {
		client.options.logger = zLogger.NewGormLogger(client.IsDebug(), 4)
	}

	// Return the client
	return client, nil
}

// IsDebug will return if debugging is enabled
func (c *Client) IsDebug() bool {
	return c.options.debug
}

// Debug will set the debug flag
func (c *Client) Debug(on bool) {
	c.options.debug = on
}

// Logger get the logger
func (c *Client) Logger() zLogger.GormLoggerInterface {
	return c.options.logger
}
