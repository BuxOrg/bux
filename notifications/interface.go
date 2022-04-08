package notifications

import (
	"context"
	"net/http"
)

// HTTPInterface is the HTTP client interface
type HTTPInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientInterface is the notification client interface
type ClientInterface interface {
	Notify(ctx context.Context, eventType EventType, model interface{}, id string) error
}
