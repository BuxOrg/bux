package notifications

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
	}
}

// WithNotifications will set the default http endpoint notifications handler
func WithNotifications(webhookEndpoint string) ClientOps {
	return func(c *clientOptions) {
		c.config.webhookEndpoint = webhookEndpoint
	}
}
