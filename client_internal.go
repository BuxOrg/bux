package bux

import (
	"context"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/cluster"
	"github.com/BuxOrg/bux/notifications"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/mrz1836/go-cachestore"
	"github.com/mrz1836/go-datastore"
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

// loadCluster will load the cluster coordinator
func (c *Client) loadCluster(ctx context.Context) (err error) {
	// Load if a custom interface was NOT provided
	if c.options.cluster.ClientInterface == nil {
		c.options.cluster.ClientInterface, err = cluster.NewClient(ctx, c.options.cluster.options...)
	}

	return
}

// loadChainstate will load chainstate configuration and start the Chainstate client
func (c *Client) loadChainstate(ctx context.Context) (err error) {
	// Load chainstate if a custom interface was NOT provided
	if c.options.chainstate.ClientInterface == nil {
		c.options.chainstate.options = append(c.options.chainstate.options, chainstate.WithUserAgent(c.UserAgent()))
		c.options.chainstate.options = append(c.options.chainstate.options, chainstate.WithHTTPClient(c.HTTPClient()))
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
	if c.options.dataStore.ClientInterface == nil {

		// Add custom array and object fields
		c.options.dataStore.options = append(
			c.options.dataStore.options,
			datastore.WithCustomFields(
				[]string{ // Array fields
					"xpub_in_ids",
					"xpub_out_ids",
				}, []string{ // Object fields
					"xpub_metadata",
					"xpub_output_value",
				},
			))

		// Add custom mongo processor
		c.options.dataStore.options = append(
			c.options.dataStore.options,
			datastore.WithCustomMongoConditionProcessor(processCustomFields),
		)

		// Add custom mongo indexes
		c.options.dataStore.options = append(
			c.options.dataStore.options,
			datastore.WithCustomMongoIndexer(getMongoIndexes),
		)

		// Load the datastore client
		c.options.dataStore.ClientInterface, err = datastore.NewClient(
			ctx, c.options.dataStore.options...,
		)
	}
	return
}

// loadNotificationClient will load the notifications client
func (c *Client) loadNotificationClient() (err error) {
	// Load notification if a custom interface was NOT provided
	if c.options.notifications.ClientInterface == nil {
		c.options.notifications.ClientInterface, err = notifications.NewClient(c.options.notifications.options...)
	}
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

// loadMonitor will load the default Monitor
//
// Cachestore is required to be loaded before this method is called
func (c *Client) loadMonitor(ctx context.Context) (err error) {
	// Check if the monitor was set by the user
	monitor := c.options.chainstate.Monitor()
	if monitor == nil {
		return // No monitor, exit!
	}

	// Create a handler and load destinations if option has been set
	handler := NewMonitorHandler(ctx, c, monitor)

	// Start the default monitor
	if err = startDefaultMonitor(ctx, c, monitor); err != nil {
		return err
	}

	lockKey := c.options.cluster.GetClusterPrefix() + lockKeyMonitorLockID
	lockID := monitor.GetLockID()
	go func() {
		var currentLock string
		for {
			if currentLock, err = c.Cachestore().WriteLockWithSecret(ctx, lockKey, lockID, defaultMonitorLockTTL); err != nil {
				// do nothing really, we just didn't get the lock
				if monitor.IsDebug() {
					monitor.Logger().Info(ctx, fmt.Sprintf("[MONITOR] failed getting lock for monitor: %s: %e", lockID, err))
				}
			}

			if lockID == currentLock {
				// Start the monitor, if not connected
				if !monitor.IsConnected() {
					if err = monitor.Start(ctx, &handler, func() {
						_, err = c.Cachestore().ReleaseLock(ctx, lockKeyMonitorLockID, lockID)
					}); err != nil {
						monitor.Logger().Error(ctx, fmt.Sprintf("[MONITOR] ERROR: failed starting monitor: %e", err))
					}
				}
			} else {
				// first close any monitor if running
				if monitor.IsConnected() {
					if err = monitor.Stop(ctx); err != nil {
						monitor.Logger().Info(ctx, fmt.Sprintf("[MONITOR] ERROR: failed stopping monitor: %e", err))
					}
				}
			}

			time.Sleep(defaultMonitorSleep)
		}
	}()

	return nil
}

// runModelMigrations will run the model Migrate() method for all models
func (c *Client) runModelMigrations(models ...interface{}) (err error) {
	// If the migrations are disabled, just return
	if c.options.dataStore.migrationDisabled {
		return nil
	}

	// Migrate the models
	d := c.Datastore()
	for _, model := range models {
		model.(ModelInterface).SetOptions(WithClient(c))
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

func (c *Client) loadPulse(ctx context.Context) (err error) {
	// Check if pulse was set by the user
	pulse := c.options.chainstate.Pulse()
	if pulse == nil {
		return
	}

	handler := NewPulseHandler(ctx, c, pulse)

	go func() {
		if !pulse.IsConnected() {
			if err = pulse.Start(ctx, &handler); err != nil {
				pulse.Logger().Error(ctx, fmt.Sprintf("[PULSE] ERROR: failed starting pulse monitor: %e", err))
			}
		}
	}()

	return nil
}
