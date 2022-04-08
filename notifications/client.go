package notifications

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
		httpClient HTTPInterface
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

	// Return the client
	return client, nil
}
