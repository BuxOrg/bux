package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mrz1836/go-logger"
)

// GetWebhookEndpoint get the configured webhook endpoint
func (c *Client) GetWebhookEndpoint() string {
	return c.options.config.webhookEndpoint
}

// Notify create a new notification
func (c *Client) Notify(ctx context.Context, modelType string, eventType EventType, model interface{}, id string) error {

	if len(c.options.config.webhookEndpoint) == 0 {
		logger.Printf("NOTIFY %s: %s - %v", eventType, id, model)
	} else {
		jsonData, err := json.Marshal(map[string]interface{}{
			"modelType": modelType,
			"eventType": eventType,
			"id":        id,
			"model":     model,
		})
		if err != nil {
			return err
		}

		var req *http.Request
		if req, err = http.NewRequestWithContext(ctx,
			http.MethodPost,
			c.options.config.webhookEndpoint,
			bytes.NewBuffer(jsonData),
		); err != nil {
			return err
		}

		var response *http.Response
		if response, err = c.options.httpClient.Do(req); err != nil {
			return err
		}
		defer func() {
			_ = response.Body.Close()
		}()

		if response.StatusCode != http.StatusOK {
			// todo queue notification for another try ...
			logger.Data(2, logger.INFO,
				fmt.Sprintf("%s: %d", "received invalid response from notification endpoint: ",
					response.StatusCode,
				),
			)
		}
	}

	return nil
}
