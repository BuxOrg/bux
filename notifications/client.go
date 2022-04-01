package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mrz1836/go-logger"
)

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
		options    *clientOptions
		httpClient HTTPInterface
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
		config *notificationsConfig // Configuration for broadcasting and other chain-state actions
	}

	// syncConfig holds all the configuration about the different sync processes
	notificationsConfig struct {
		webhookEndpoint string
	}
)

// NewClient creates a new client for notifications
func NewClient(opts ...ClientOps) (ClientInterface, error) {

	// Create a new client with defaults
	client := &Client{
		options:    defaultClientOptions(),
		httpClient: http.DefaultClient,
	}

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(client.options)
	}

	// Return the client
	return client, nil
}

// Notify create a new notification
func (c *Client) Notify(ctx context.Context, eventType EventType, model interface{}, id string) error {
	if c.options.config.webhookEndpoint == "" {
		fmt.Printf("NOTIFY %s: %s - %v", eventType, id, model)
	} else {
		jsonData, err := json.Marshal(map[string]interface{}{
			"event": eventType,
			"id":    id,
			"model": model,
		})
		if err != nil {
			return err
		}

		var req *http.Request
		req, err = http.NewRequestWithContext(ctx,
			http.MethodPost,
			c.options.config.webhookEndpoint,
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return err
		}

		var response *http.Response
		response, err = c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			// todo queue notification for another try ...
			logger.Data(2, logger.INFO,
				fmt.Sprintf("%s: %d", "received invalid response from notification endpoint: ", response.StatusCode),
			)
		}
	}

	return nil
}
