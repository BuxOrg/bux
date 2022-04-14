package notifications

import "github.com/BuxOrg/bux/logger"

// EventType event types thrown in Bux
type EventType string

const (
	// EventTypeCreate when a new model is created
	EventTypeCreate EventType = "create"

	// EventTypeRead when a new model is read
	EventTypeRead EventType = "read"

	// EventTypeUpdate when a new model is updated
	EventTypeUpdate EventType = "update"

	// EventTypeDelete when a new model is deleted
	EventTypeDelete EventType = "delete"
)

type (

	// Client is the client (configuration)
	Client struct {
		options *clientOptions
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
		config     *notificationsConfig // Configuration for broadcasting and other chain-state actions
		debug      bool
		httpClient HTTPInterface
		logger     logger.Interface
	}

	// syncConfig holds all the configuration about the different notifications
	notificationsConfig struct {
		webhookEndpoint string
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
		client.options.logger = logger.NewLogger(client.IsDebug())
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
func (c *Client) Logger() logger.Interface {
	return c.options.logger
}
