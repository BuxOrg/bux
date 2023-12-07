package notifications

import (
	"context"
	"github.com/rs/zerolog"
	"net/http"
)

// HTTPInterface is the HTTP client interface
type HTTPInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientInterface is the notification client interface
type ClientInterface interface {
	Debug(on bool)
	GetWebhookEndpoint() string
	IsDebug() bool
	Logger() *zerolog.Logger
	Notify(ctx context.Context, modelType string, eventType EventType, model interface{}, id string) error
}
