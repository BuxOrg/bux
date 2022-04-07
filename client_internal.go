package bux

import (
	"context"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
)

// loadCache will load caching configuration and start the Cachestore client
func (c *Client) loadCache(ctx context.Context) (err error) {

	// Load if a custom interface was NOT provided
	if c.options.cacheStore.ClientInterface == nil {
		c.options.cacheStore.ClientInterface, err = cachestore.NewClient(ctx, c.options.cacheStore.options...)
	}
	return
}

// loadChainstate will load chainstate configuration and start the Chainstate client
func (c *Client) loadChainstate(ctx context.Context) (err error) {

	// Load chainstate if a custom interface was NOT provided
	if c.options.chainstate.ClientInterface == nil {
		c.options.chainstate.options = append(c.options.chainstate.options, chainstate.WithUserAgent(c.UserAgent()))
		c.options.chainstate.ClientInterface, err = chainstate.NewClient(ctx, c.options.chainstate.options...)
	}
	return
}

// loadDatastore will load the Datastore and start the Datastore client
//
// NOTE: this will run database migrations if the options was set
func (c *Client) loadDatastore(ctx context.Context) (err error) {

	// Add the models to migrate (after loading the client options)
	if len(c.options.models.migrateModelNames) > 0 {
		c.options.dataStore.options = append(
			c.options.dataStore.options,
			datastore.WithAutoMigrate(c.options.models.migrateModels...),
		)
	}

	// Load client (runs ALL options, IE: auto migrate models)
	c.options.dataStore.ClientInterface, err = datastore.NewClient(
		ctx, c.options.dataStore.options...,
	)
	return
}

// loadPaymailClient will load the Paymail client
func (c *Client) loadPaymailClient() (err error) {
	// Only load if it's not set (the client can be overloaded)
	if c.options.paymail.client == nil {
		c.options.paymail.client, err = paymail.NewClient()
	}
	return
}

// loadTaskmanager will load the TaskManager and start the TaskManager client
func (c *Client) loadTaskmanager(ctx context.Context) (err error) {

	// Load if a custom interface was NOT provided
	if c.options.taskManager.ClientInterface == nil {
		c.options.taskManager.ClientInterface, err = taskmanager.NewClient(
			ctx, c.options.taskManager.options...,
		)
	}
	return
}

// runModelMigrations will run the model Migrate() method for all models
func (c *Client) runModelMigrations(models ...interface{}) (err error) {
	d := c.Datastore()
	for _, model := range models {
		if err = model.(ModelInterface).Migrate(d); err != nil {
			return
		}
	}
	return
}

// runModelRegisterTasks will run the model RegisterTasks() method for all models
func (c *Client) runModelRegisterTasks(models ...interface{}) (err error) {
	for _, model := range models {
		model.(ModelInterface).SetOptions(c.DefaultModelOptions()...)
		if err = model.(ModelInterface).RegisterTasks(); err != nil {
			return
		}
	}
	return
}

// registerAllTasks will register all tasks for all models
func (c *Client) registerAllTasks() error {
	c.Taskmanager().ResetCron()
	return c.runModelRegisterTasks(c.options.models.models...)
}

// loadDefaultPaymailConfig will load the default paymail server configuration
func (c *Client) loadDefaultPaymailConfig() (err error) {

	// Default FROM paymail
	if len(c.options.paymail.serverConfig.DefaultFromPaymail) == 0 {
		c.options.paymail.serverConfig.DefaultFromPaymail = defaultSenderPaymail
	}

	// Default note for address resolution
	if len(c.options.paymail.serverConfig.DefaultNote) == 0 {
		c.options.paymail.serverConfig.DefaultNote = defaultAddressResolutionPurpose
	}

	// Set default options if none are found
	if len(c.options.paymail.serverConfig.options) == 0 {
		c.options.paymail.serverConfig.options = append(c.options.paymail.serverConfig.options,
			server.WithP2PCapabilities(),
			server.WithDomainValidationDisabled(),
		)
	}

	// Create the paymail configuration using the client and default service provider
	c.options.paymail.serverConfig.Configuration, err = server.NewConfig(
		&PaymailDefaultServiceProvider{client: c},
		c.options.paymail.serverConfig.options...,
	)
	return
}
