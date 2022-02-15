package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/logger"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/utils"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
	glogger "gorm.io/gorm/logger"
)

type (

	// Client is the bux client & options
	Client struct {
		options *clientOptions
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
		cacheStore  *cacheStoreOptions  // Configuration options for Cachestore (ristretto, redis, etc.)
		chainstate  *chainstateOptions  // Configuration options for Chainstate (broadcast, sync, etc.)
		dataStore   *dataStoreOptions   // Configuration options for the DataStore (MySQL, etc.)
		debug       bool                // If the client is in debug mode
		itc         bool                // (Incoming Transactions Check) True will check incoming transactions via Miners (real-world)
		iuc         bool                // (Input UTXO Check) True will check input utxos when saving transactions
		logger      glogger.Interface   // Internal logging
		models      *modelOptions       // Configuration options for the loaded models
		newRelic    *newRelicOptions    // Configuration options for NewRelic
		paymail     *paymailOptions     // Paymail options & client
		taskManager *taskManagerOptions // Configuration options for the TaskManager (TaskQ, etc.)
		userAgent   string              // User agent for all outgoing requests
	}

	// chainstateOptions holds the chainstate configuration and client
	chainstateOptions struct {
		chainstate.ClientInterface                        // Client for Chainstate
		options                    []chainstate.ClientOps // List of options
	}

	// cacheStoreOptions holds the cache configuration and client
	cacheStoreOptions struct {
		cachestore.ClientInterface                        // Client for Cachestore
		options                    []cachestore.ClientOps // List of options
	}

	// dataStoreOptions holds the data storage configuration and client
	dataStoreOptions struct {
		datastore.ClientInterface                       // Client for Datastore
		options                   []datastore.ClientOps // List of options
	}

	// modelOptions holds the model configuration
	modelOptions struct {
		migrateModelNames []string      // List of models for migration
		migrateModels     []interface{} // Models for migrations
		modelNames        []string      // List of all models
		models            []interface{} // Models for use in this engine
	}

	// newRelicOptions holds the configuration for NewRelic
	newRelicOptions struct {
		app     *newrelic.Application // NewRelic client application (if enabled)
		enabled bool                  // If NewRelic is enabled for deep Transaction tracing
	}

	// paymailOptions holds the configuration for Paymail
	paymailOptions struct {
		client       paymail.ClientInterface // Paymail client for communicating with Paymail providers
		serverConfig *paymailServerOptions   // Server configuration if Paymail is enabled
	}

	// paymailServerOptions is the options for the Paymail server
	paymailServerOptions struct {
		*server.Configuration        // Server configuration if Paymail is enabled
		DefaultFromPaymail    string // IE: from@domain.com
		DefaultNote           string // IE: some note for address resolution
	}

	// taskManagerOptions holds the configuration for taskmanager
	taskManagerOptions struct {
		taskmanager.ClientInterface                          // Client for TaskManager
		cronTasks                   map[string]time.Duration // List of tasks and period times (IE: task_name 30*time.Minute = @every 30m)
		options                     []taskmanager.ClientOps  // List of options
	}
)

// NewClient creates a new client for all bux functionality
//
// If no options are given, it will use the defaultClientOptions()
// ctx may contain a NewRelic txn (or one will be created)
func NewClient(ctx context.Context, opts ...ClientOps) (ClientInterface, error) {

	// Create a new client with defaults
	client := &Client{options: defaultClientOptions()}

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(client.options)
	}

	// Use NewRelic if it's enabled (use existing txn if found on ctx)
	ctx = client.GetOrStartTxn(ctx, "new_client")

	// Load the cachestore client
	if err := client.loadCache(ctx); err != nil {
		return nil, err
	}

	// Load the datastore (automatically migrate models)
	if err := client.loadDatastore(ctx); err != nil {
		return nil, err
	}

	// Run custom model datastore migrations
	if err := client.runModelMigrations(
		client.options.models.migrateModels...,
	); err != nil {
		return nil, err
	}

	// Load the chainstate client
	if err := client.loadChainstate(ctx); err != nil {
		return nil, err
	}

	// Load the Paymail client (if client does not exist)
	if err := client.loadPaymailClient(); err != nil {
		return nil, err
	}

	// Load the taskmanager (automatically start consumers and tasks)
	if err := client.loadTaskmanager(ctx); err != nil {
		return nil, err
	}

	// Load all model tasks
	if err := client.registerAllTasks(); err != nil {
		return nil, err
	}

	// Set logger (if not set by user)
	if client.options.logger == nil {
		client.options.logger = logger.NewLogger(client.IsDebug())
	}

	// Return the client
	return client, nil
}

// AddModels will add additional models to the client
func (c *Client) AddModels(ctx context.Context, autoMigrate bool, models ...interface{}) error {

	// Store the models locally in the client
	c.options.addModels(modelList, models...)

	// Should we migrate the models?
	if autoMigrate {

		// Ensure we have a datastore
		d := c.Datastore()
		if d == nil {
			return ErrDatastoreRequired
		}

		// Apply the database migration with the new models
		if err := d.AutoMigrateDatabase(ctx, models...); err != nil {
			return err
		}

		// Add to the list
		c.options.addModels(migrateList, models...)

		// Run model migrations
		if err := c.runModelMigrations(models...); err != nil {
			return err
		}
	}

	// Register all tasks (again)
	return c.registerAllTasks()
}

// Cachestore will return the Cachestore IF: exists and is enabled
func (c *Client) Cachestore() cachestore.ClientInterface {
	if c.options.cacheStore != nil && c.options.cacheStore.ClientInterface != nil {
		return c.options.cacheStore.ClientInterface
	}
	return nil
}

// Chainstate will return the Chainstate service IF: exists and is enabled
func (c *Client) Chainstate() chainstate.ClientInterface {
	if c.options.chainstate != nil && c.options.chainstate.ClientInterface != nil {
		return c.options.chainstate.ClientInterface
	}
	return nil
}

// Close will safely close any open connections (cache, datastore, etc.)
func (c *Client) Close(ctx context.Context) error {

	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("close_all").End()
	}

	// Close Cachestore
	cs := c.Cachestore()
	if cs != nil {
		cs.Close(ctx)
		c.options.cacheStore.ClientInterface = nil
	}

	// Close Chainstate
	ch := c.Chainstate()
	if ch != nil {
		ch.Close(ctx)
		c.options.chainstate.ClientInterface = nil
	}

	// Close Datastore
	ds := c.Datastore()
	if ds != nil {
		if err := ds.Close(ctx); err != nil {
			return err
		}
		c.options.dataStore.ClientInterface = nil
	}

	// Close Taskmanager
	tm := c.Taskmanager()
	if tm != nil {
		if err := tm.Close(ctx); err != nil {
			return err
		}
		c.options.taskManager.ClientInterface = nil
	}
	return nil
}

// Datastore will return the Datastore if it exists
func (c *Client) Datastore() datastore.ClientInterface {
	if c.options.dataStore != nil && c.options.dataStore.ClientInterface != nil {
		return c.options.dataStore.ClientInterface
	}
	return nil
}

// Logger will return the Logger if it exists
func (c *Client) Logger() glogger.Interface {
	return c.options.logger
}

// Debug will toggle the debug mode (for all resources)
func (c *Client) Debug(on bool) {

	// Set the flag on the current client
	c.options.debug = on

	// Set debugging on the Cachestore
	cs := c.Cachestore()
	if cs != nil {
		cs.Debug(on)
	}

	// Set debugging on the Chainstate
	ch := c.Chainstate()
	if ch != nil {
		ch.Debug(on)
	}

	// Set debugging on the Datastore
	ds := c.Datastore()
	if ds != nil {
		ds.Debug(on)
	}

	// Set debugging on the Taskmanager
	tm := c.Taskmanager()
	if tm != nil {
		tm.Debug(on)
	}
}

// EnableNewRelic will enable NewRelic tracing
func (c *Client) EnableNewRelic() {
	if c.options.newRelic != nil && c.options.newRelic.app != nil {
		c.options.newRelic.enabled = true
	}
}

// GetOrStartTxn will check for an existing NewRelic transaction, if not found, it will make a new transaction
func (c *Client) GetOrStartTxn(ctx context.Context, name string) context.Context {
	if c.IsNewRelicEnabled() && c.options.newRelic.app != nil {
		txn := newrelic.FromContext(ctx)
		if txn == nil {
			txn = c.options.newRelic.app.StartTransaction(name)
		}
		ctx = newrelic.NewContext(ctx, txn)
	}
	return ctx
}

// GetFeeUnit get the fee from a miner
// todo: move into it's own Service / package
func (c *Client) GetFeeUnit(_ context.Context, _ string) *utils.FeeUnit {
	return defaultFee
	/*
		func (c *Client) GetFeeUnit(ctx context.Context, useMiner string) *utils.FeeUnit {
			if useMiner == "" {
				useMiner = minercraft.MinerTaal
			}
			cacheKey := "draft-transaction-fee-" + useMiner

			var fee *utils.FeeUnit
			err := c.Cachestore().GetModel(ctx, cacheKey, fee)
			if err == nil && fee != nil {
				return fee
			}

			var client *minercraft.Client
			if client, err = minercraft.NewClient(nil, nil, nil); err != nil {
				log.Printf("error occurred creating minercraft client: %s", err.Error())
				return defaultFee
			}

			miner := client.MinerByName(useMiner)
			var response *minercraft.FeeQuoteResponse
			if response, err = client.FeeQuote(context.Background(), miner); err != nil {
				log.Printf("error occurred getting fee quote: %s", err.Error())
				return defaultFee
			}

			for _, quote := range response.Quote.Fee {
				if quote.FeeType == "standard" && quote.MiningFee.Satoshis > 0 {
					fee = &utils.FeeUnit{
						Satoshis: quote.MiningFee.Satoshis,
						Bytes:    quote.MiningFee.Bytes,
					}
					break
				}
			}

			err = c.Cachestore().SetModel(ctx, cacheKey, fee)
			if err != nil {
				log.Printf("error occurred caching fee quote: %s", err.Error())
			}

			return fee
		}
	*/
}

// GetTaskPeriod will return the period for a given task name
func (c *Client) GetTaskPeriod(name string) time.Duration {
	if d, ok := c.options.taskManager.cronTasks[name]; ok {
		return d
	}
	return 0
}

// IsDebug will return the debug flag (bool)
func (c *Client) IsDebug() bool {
	return c.options.debug
}

// IsNewRelicEnabled will return the flag (bool)
func (c *Client) IsNewRelicEnabled() bool {
	return c.options.newRelic.enabled
}

// IsITCEnabled will return the flag (bool)
func (c *Client) IsITCEnabled() bool {
	return c.options.itc
}

// IsIUCEnabled will return the flag (bool)
func (c *Client) IsIUCEnabled() bool {
	return c.options.iuc
}

// ModifyTaskPeriod will modify a cron task's duration period from the default
func (c *Client) ModifyTaskPeriod(name string, period time.Duration) error {

	// Basic validation on parameters
	if len(name) == 0 {
		return taskmanager.ErrMissingTaskName
	} else if period <= 0 {
		return taskmanager.ErrInvalidTaskDuration
	}

	// Ensure task manager has been loaded
	if c.Taskmanager() == nil || c.options.taskManager.cronTasks == nil {
		return ErrTaskManagerNotLoaded
	} else if len(c.options.taskManager.cronTasks) == 0 {
		return taskmanager.ErrNoTasksFound
	}

	// Check for the task
	if d, ok := c.options.taskManager.cronTasks[name]; !ok {
		return taskmanager.ErrTaskNotFound
	} else if d == period {
		return nil
	}

	// Set the new period on the client
	c.options.taskManager.cronTasks[name] = period

	// register all tasks again (safely override)
	return c.registerAllTasks()
}

// Taskmanager will return the Taskmanager if it exists
func (c *Client) Taskmanager() taskmanager.ClientInterface {
	if c.options.taskManager != nil && c.options.taskManager.ClientInterface != nil {
		return c.options.taskManager.ClientInterface
	}
	return nil
}

// UserAgent will return the user agent
func (c *Client) UserAgent() string {
	return c.options.userAgent
}

// Version will return the version
func (c *Client) Version() string {
	return version
}
