package notifications

// EventType event types thrown in Bux
type EventType string

const (
	// EventTypeDestinationCreate when a new destination is created
	EventTypeDestinationCreate EventType = "destination_create"

	// EventTypeTransactionCreate when a new transaction is created
	EventTypeTransactionCreate EventType = "transaction_create"
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
