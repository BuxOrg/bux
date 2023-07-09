package bux

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/cluster"
	"github.com/BuxOrg/bux/notifications"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/coocood/freecache"
	"github.com/go-redis/redis/v8"
	"github.com/mrz1836/go-cache"
	"github.com/mrz1836/go-cachestore"
	"github.com/mrz1836/go-datastore"
	zLogger "github.com/mrz1836/go-logger"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-minercraft"
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
	taskq "github.com/vmihailenco/taskq/v3"
	"go.mongodb.org/mongo-driver/mongo"
)

// ClientOps allow functional options to be supplied that overwrite default client options.
type ClientOps func(c *clientOptions)

// defaultClientOptions will return an clientOptions struct with the default settings
//
// Useful for starting with the default and then modifying as needed
func defaultClientOptions() *clientOptions {

	// Set the default options
	return &clientOptions{

		// Incoming Transaction Checker (lookup external tx via miner for validity)
		itc: true,

		// By default check input utxos (unless disabled by the user)
		iuc: true,

		// Blank chainstate config
		chainstate: &chainstateOptions{
			ClientInterface:  nil,
			options:          []chainstate.ClientOps{},
			broadcasting:     true, // Enabled by default for new users
			broadcastInstant: true, // Enabled by default for new users
			paymailP2P:       true, // Enabled by default for new users
			syncOnChain:      true, // Enabled by default for new users
		},

		cluster: &clusterOptions{
			options: []cluster.ClientOps{},
		},

		// Blank cache config
		cacheStore: &cacheStoreOptions{
			ClientInterface: nil,
			options:         []cachestore.ClientOps{},
		},

		// Blank Datastore config
		dataStore: &dataStoreOptions{
			ClientInterface: nil,
			options:         []datastore.ClientOps{},
		},

		// Default http client
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},

		// Blank model options (use the Base models)
		models: &modelOptions{
			modelNames:        modelNames(BaseModels...),
			models:            BaseModels,
			migrateModelNames: nil,
			migrateModels:     nil,
		},

		// Blank NewRelic config
		newRelic: &newRelicOptions{},

		// Blank notifications config
		notifications: &notificationsOptions{
			ClientInterface: nil,
			webhookEndpoint: "",
		},

		// Blank Paymail config
		paymail: &paymailOptions{
			client: nil,
			serverConfig: &PaymailServerOptions{
				Configuration: nil,
				options:       []server.ConfigOps{},
			},
		},

		// Blank TaskManager config
		taskManager: &taskManagerOptions{
			ClientInterface: nil,
			cronTasks: map[string]time.Duration{
				ModelDestination.String() + "_monitor":                    taskIntervalMonitorCheck,
				ModelDraftTransaction.String() + "_clean_up":              taskIntervalDraftCleanup,
				ModelIncomingTransaction.String() + "_process":            taskIntervalProcessIncomingTxs,
				ModelSyncTransaction.String() + "_" + syncActionBroadcast: taskIntervalSyncActionBroadcast,
				ModelSyncTransaction.String() + "_" + syncActionP2P:       taskIntervalSyncActionP2P,
				ModelSyncTransaction.String() + "_" + syncActionSync:      taskIntervalSyncActionSync,
			},
		},

		// Default user agent
		userAgent: defaultUserAgent,
	}
}

// modelNames will take a list of models and return the list of names
func modelNames(models ...interface{}) (names []string) {
	for _, modelInterface := range models {
		names = append(names, modelInterface.(ModelInterface).Name())
	}
	return
}

// modelExists will return true if the model is found
func (o *clientOptions) modelExists(modelName, list string) bool {
	m := o.models.modelNames
	if list == migrateList {
		m = o.models.migrateModelNames
	}
	for _, name := range m {
		if strings.EqualFold(name, modelName) {
			return true
		}
	}
	return false
}

// addModel will add the model if it does not exist already (load once)
func (o *clientOptions) addModel(model interface{}, list string) {
	name := model.(ModelInterface).Name()
	if !o.modelExists(name, list) {
		if list == migrateList {
			o.models.migrateModelNames = append(o.models.migrateModelNames, name)
			o.models.migrateModels = append(o.models.migrateModels, model)
			return
		}
		o.models.modelNames = append(o.models.modelNames, name)
		o.models.models = append(o.models.models, model)
	}
}

// addModels will add the models if they do not exist already (load once)
func (o *clientOptions) addModels(list string, models ...interface{}) {
	for _, modelInterface := range models {
		o.addModel(modelInterface, list)
	}
}

// DefaultModelOptions will set any default model options (from Client options->model)
func (c *Client) DefaultModelOptions(opts ...ModelOps) []ModelOps {

	// Set the Client from the bux.Client onto the model
	opts = append(opts, WithClient(c))

	// Set the encryption key (if found)
	opts = append(opts, WithEncryptionKey(c.options.encryptionKey))

	// Return the new options
	return opts
}

// -----------------------------------------------------------------
// GENERAL
// -----------------------------------------------------------------

// WithUserAgent will overwrite the default useragent
func WithUserAgent(userAgent string) ClientOps {
	return func(c *clientOptions) {
		if len(userAgent) > 0 {
			c.userAgent = userAgent
		}
	}
}

// WithNewRelic will set the NewRelic application client
func WithNewRelic(app *newrelic.Application) ClientOps {
	return func(c *clientOptions) {
		// Disregard if the app is nil
		if app == nil {
			return
		}

		// Set the app
		c.newRelic.app = app

		// Enable New relic on other services
		c.cacheStore.options = append(c.cacheStore.options, cachestore.WithNewRelic())
		c.chainstate.options = append(c.chainstate.options, chainstate.WithNewRelic())
		c.dataStore.options = append(c.dataStore.options, datastore.WithNewRelic())
		c.taskManager.options = append(c.taskManager.options, taskmanager.WithNewRelic())
		// c.notifications.options = append(c.notifications.options, notifications.WithNewRelic())

		// Enable the service
		c.newRelic.enabled = true
	}
}

// WithDebugging will set debugging in any applicable configuration
func WithDebugging() ClientOps {
	return func(c *clientOptions) {
		c.debug = true

		// Enable debugging on other services
		c.cacheStore.options = append(c.cacheStore.options, cachestore.WithDebugging())
		c.chainstate.options = append(c.chainstate.options, chainstate.WithDebugging())
		c.dataStore.options = append(c.dataStore.options, datastore.WithDebugging())
		c.notifications.options = append(c.notifications.options, notifications.WithDebugging())
		c.taskManager.options = append(c.taskManager.options, taskmanager.WithDebugging())
	}
}

// WithEncryption will set the encryption key and encrypt values using this key
func WithEncryption(key string) ClientOps {
	return func(c *clientOptions) {
		if len(key) > 0 {
			c.encryptionKey = key
		}
	}
}

// WithModels will add additional models (will NOT migrate using datastore)
//
// Pointers of structs (IE: &models.Xpub{})
func WithModels(models ...interface{}) ClientOps {
	return func(c *clientOptions) {
		if len(models) > 0 {
			c.addModels(modelList, models...)
		}
	}
}

// WithITCDisabled will disable (ITC) incoming transaction checking
func WithITCDisabled() ClientOps {
	return func(c *clientOptions) {
		c.itc = false
	}
}

// WithIUCDisabled will disable checking the input utxos
func WithIUCDisabled() ClientOps {
	return func(c *clientOptions) {
		c.iuc = false
	}
}

// WithImportBlockHeaders will import block headers on startup
func WithImportBlockHeaders(importBlockHeadersURL string) ClientOps {
	return func(c *clientOptions) {
		if len(importBlockHeadersURL) > 0 {
			c.importBlockHeadersURL = importBlockHeadersURL
		}
	}
}

// WithHTTPClient will set the custom http interface
func WithHTTPClient(httpClient HTTPInterface) ClientOps {
	return func(c *clientOptions) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithLogger will set the custom logger interface
func WithLogger(customLogger zLogger.GormLoggerInterface) ClientOps {
	return func(c *clientOptions) {
		if customLogger != nil {
			c.logger = customLogger

			// Enable the logger on all services
			c.cacheStore.options = append(c.cacheStore.options, cachestore.WithLogger(c.logger))
			c.chainstate.options = append(c.chainstate.options, chainstate.WithLogger(c.logger))
			c.dataStore.options = append(c.dataStore.options, datastore.WithLogger(&datastore.DatabaseLogWrapper{GormLoggerInterface: c.logger}))
			c.taskManager.options = append(c.taskManager.options, taskmanager.WithLogger(c.logger))
			c.notifications.options = append(c.notifications.options, notifications.WithLogger(c.logger))
		}
	}
}

// -----------------------------------------------------------------
// CACHESTORE
// -----------------------------------------------------------------

// WithCustomCachestore will set the cachestore
func WithCustomCachestore(cacheStore cachestore.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if cacheStore != nil {
			c.cacheStore.ClientInterface = cacheStore
		}
	}
}

// WithFreeCache will set the cache client for both Read & Write clients
func WithFreeCache() ClientOps {
	return func(c *clientOptions) {
		c.cacheStore.options = append(c.cacheStore.options, cachestore.WithFreeCache())
	}
}

// WithFreeCacheConnection will set the cache client to an active FreeCache connection
func WithFreeCacheConnection(client *freecache.Cache) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.cacheStore.options = append(
				c.cacheStore.options,
				cachestore.WithFreeCacheConnection(client),
			)
		}
	}
}

// WithRedis will set the redis cache client for both Read & Write clients
//
// This will load new redis connections using the given parameters
func WithRedis(config *cachestore.RedisConfig) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.cacheStore.options = append(c.cacheStore.options, cachestore.WithRedis(config))
		}
	}
}

// WithRedisConnection will set the cache client to an active redis connection
func WithRedisConnection(activeClient *cache.Client) ClientOps {
	return func(c *clientOptions) {
		if activeClient != nil {
			c.cacheStore.options = append(
				c.cacheStore.options,
				cachestore.WithRedisConnection(activeClient),
			)
		}
	}
}

// -----------------------------------------------------------------
// DATASTORE
// -----------------------------------------------------------------

// WithCustomDatastore will set the datastore
func WithCustomDatastore(dataStore datastore.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if dataStore != nil {
			c.dataStore.ClientInterface = dataStore
		}
	}
}

// WithAutoMigrate will enable auto migrate database mode (given models)
//
// Pointers of structs (IE: &models.Xpub{})
func WithAutoMigrate(migrateModels ...interface{}) ClientOps {
	return func(c *clientOptions) {
		if len(migrateModels) > 0 {
			c.addModels(modelList, migrateModels...)
			c.addModels(migrateList, migrateModels...)
		}
	}
}

// WithMigrationDisabled will disable all migrations from running in the Datastore
func WithMigrationDisabled() ClientOps {
	return func(c *clientOptions) {
		c.dataStore.migrationDisabled = true
	}
}

// WithSQLite will set the Datastore to use SQLite
func WithSQLite(config *datastore.SQLiteConfig) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.dataStore.options = append(c.dataStore.options, datastore.WithSQLite(config))
		}
	}
}

// WithSQL will set the datastore to use the SQL config
func WithSQL(engine datastore.Engine, config *datastore.SQLConfig) ClientOps {
	return func(c *clientOptions) {
		if config != nil && !engine.IsEmpty() {
			c.dataStore.options = append(
				c.dataStore.options,
				datastore.WithSQL(engine, []*datastore.SQLConfig{config}),
			)
		}
	}
}

// WithSQLConfigs will load multiple connections (replica & master)
func WithSQLConfigs(engine datastore.Engine, configs []*datastore.SQLConfig) ClientOps {
	return func(c *clientOptions) {
		if len(configs) > 0 && !engine.IsEmpty() {
			c.dataStore.options = append(
				c.dataStore.options,
				datastore.WithSQL(engine, configs),
			)
		}
	}
}

// WithSQLConnection will set the Datastore to an existing connection for MySQL or PostgreSQL
func WithSQLConnection(engine datastore.Engine, sqlDB *sql.DB, tablePrefix string) ClientOps {
	return func(c *clientOptions) {
		if sqlDB != nil && !engine.IsEmpty() {
			c.dataStore.options = append(
				c.dataStore.options,
				datastore.WithSQLConnection(engine, sqlDB, tablePrefix),
			)
		}
	}
}

// WithMongoDB will set the Datastore to use MongoDB
func WithMongoDB(config *datastore.MongoDBConfig) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.dataStore.options = append(c.dataStore.options, datastore.WithMongo(config))
		}
	}
}

// WithMongoConnection will set the Datastore to an existing connection for MongoDB
func WithMongoConnection(database *mongo.Database, tablePrefix string) ClientOps {
	return func(c *clientOptions) {
		if database != nil {
			c.dataStore.options = append(
				c.dataStore.options,
				datastore.WithMongoConnection(database, tablePrefix),
			)
		}
	}
}

// -----------------------------------------------------------------
// PAYMAIL
// -----------------------------------------------------------------

// WithPaymailClient will set a custom paymail client
func WithPaymailClient(client paymail.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.paymail.client = client
		}
	}
}

// WithPaymailSupport will set the configuration for Paymail support (as a server)
func WithPaymailSupport(domains []string, defaultFromPaymail, defaultNote string,
	domainValidation, senderValidation bool) ClientOps {
	return func(c *clientOptions) {

		// Add generic capabilities
		c.paymail.serverConfig.options = append(c.paymail.serverConfig.options, server.WithP2PCapabilities())

		// Add each domain
		for _, domain := range domains {
			c.paymail.serverConfig.options = append(c.paymail.serverConfig.options, server.WithDomain(domain))
		}

		// Set the sender validation
		if senderValidation {
			c.paymail.serverConfig.options = append(c.paymail.serverConfig.options, server.WithSenderValidation())
		}

		// Domain validation
		if !domainValidation {
			c.paymail.serverConfig.options = append(c.paymail.serverConfig.options, server.WithDomainValidationDisabled())
		}

		// Add default values
		if len(defaultFromPaymail) > 0 {
			c.paymail.serverConfig.DefaultFromPaymail = defaultFromPaymail
		}
		if len(defaultNote) > 0 {
			c.paymail.serverConfig.DefaultNote = defaultNote
		}

		// Add the paymail_address model in bux
		c.addModels(migrateList, newPaymail(""))
	}
}

// WithPaymailServerConfig will set the custom server configuration for Paymail
//
// This will allow overriding the Configuration.actions (paymail service provider)
func WithPaymailServerConfig(config *server.Configuration, defaultFromPaymail, defaultNote string) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.paymail.serverConfig.Configuration = config
		}
		if len(defaultFromPaymail) > 0 {
			c.paymail.serverConfig.DefaultFromPaymail = defaultFromPaymail
		}
		if len(defaultNote) > 0 {
			c.paymail.serverConfig.DefaultNote = defaultNote
		}

		// Add the paymail_address model in bux
		c.addModels(migrateList, newPaymail(""))
	}
}

// -----------------------------------------------------------------
// TASK MANAGER
// -----------------------------------------------------------------

// WithCustomTaskManager will set the taskmanager
func WithCustomTaskManager(taskManager taskmanager.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if taskManager != nil {
			c.taskManager.ClientInterface = taskManager
		}
	}
}

// WithTaskQ will set the task manager to use TaskQ & in-memory
func WithTaskQ(config *taskq.QueueOptions, factory taskmanager.Factory) ClientOps {
	return func(c *clientOptions) {
		if config != nil {
			c.taskManager.options = append(
				c.taskManager.options,
				taskmanager.WithTaskQ(config, factory),
			)
		}
	}
}

// WithTaskQUsingRedis will set the task manager to use TaskQ & Redis
func WithTaskQUsingRedis(config *taskq.QueueOptions, redisOptions *redis.Options) ClientOps {
	return func(c *clientOptions) {
		if config != nil {

			// Create a new redis client
			if config.Redis == nil {

				// Remove prefix if found
				redisOptions.Addr = strings.Replace(redisOptions.Addr, cachestore.RedisPrefix, "", -1)
				config.Redis = redis.NewClient(redisOptions)
			}

			c.taskManager.options = append(
				c.taskManager.options,
				taskmanager.WithTaskQ(config, taskmanager.FactoryRedis),
			)
		}
	}
}

// WithCronService will set the custom cron service provider
func WithCronService(cronService taskmanager.CronService) ClientOps {
	return func(c *clientOptions) {
		if cronService != nil && c.taskManager != nil {
			c.taskManager.options = append(c.taskManager.options, taskmanager.WithCronService(cronService))
		}
	}
}

// -----------------------------------------------------------------
// CLUSTER
// -----------------------------------------------------------------

// WithClusterRedis will set the cluster coordinator to use redis
func WithClusterRedis(redisOptions *redis.Options) ClientOps {
	return func(c *clientOptions) {
		if redisOptions != nil {
			c.cluster.options = append(c.cluster.options, cluster.WithRedis(redisOptions))
		}
	}
}

// WithClusterKeyPrefix will set the cluster key prefix to use for all keys in the cluster coordinator
func WithClusterKeyPrefix(prefix string) ClientOps {
	return func(c *clientOptions) {
		if prefix != "" {
			c.cluster.options = append(c.cluster.options, cluster.WithKeyPrefix(prefix))
		}
	}
}

// WithClusterClient will set the cluster options on the client
func WithClusterClient(clusterClient cluster.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if clusterClient != nil {
			c.cluster.ClientInterface = clusterClient
		}
	}
}

// -----------------------------------------------------------------
// CHAIN-STATE
// -----------------------------------------------------------------

// WithCustomChainstate will set the chainstate
func WithCustomChainstate(chainState chainstate.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if chainState != nil {
			c.chainstate.ClientInterface = chainState
		}
	}
}

// WithChainstateOptions will set chainstate defaults
func WithChainstateOptions(broadcasting, broadcastInstant, paymailP2P, syncOnChain bool) ClientOps {
	return func(c *clientOptions) {
		c.chainstate.broadcasting = broadcasting
		c.chainstate.broadcastInstant = broadcastInstant
		c.chainstate.paymailP2P = paymailP2P
		c.chainstate.syncOnChain = syncOnChain
	}
}

// WithBroadcastMiners will set a list of miners for broadcasting
func WithBroadcastMiners(miners []*chainstate.Miner) ClientOps {
	return func(c *clientOptions) {
		if len(miners) > 0 {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithBroadcastMiners(miners))
		}
	}
}

// WithQueryMiners will set a list of miners for querying transactions
func WithQueryMiners(miners []*chainstate.Miner) ClientOps {
	return func(c *clientOptions) {
		if len(miners) > 0 {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithQueryMiners(miners))
		}
	}
}

// WithWhatsOnChainAPIKey will set the API key
func WithWhatsOnChainAPIKey(apiKey string) ClientOps {
	return func(c *clientOptions) {
		if len(apiKey) > 0 {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithWhatsOnChainAPIKey(apiKey))
		}
	}
}

// WithNowNodesAPIKey will set the API key
func WithNowNodesAPIKey(apiKey string) ClientOps {
	return func(c *clientOptions) {
		if len(apiKey) > 0 {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithNowNodesAPIKey(apiKey))
		}
	}
}

// WithExcludedProviders will set a list of excluded providers
func WithExcludedProviders(providers []string) ClientOps {
	return func(c *clientOptions) {
		if len(providers) > 0 {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithExcludedProviders(providers))
		}
	}
}

// WithMonitoring will create a new monitorConfig interface with the given options
func WithMonitoring(ctx context.Context, monitorOptions *chainstate.MonitorOptions) ClientOps {
	return func(c *clientOptions) {
		if monitorOptions != nil {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithMonitoring(ctx, monitorOptions))
		}
	}
}

// WithMonitoringInterface will set the interface to use for monitoring the blockchain
func WithMonitoringInterface(monitor chainstate.MonitorService) ClientOps {
	return func(c *clientOptions) {
		if monitor != nil {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithMonitoringInterface(monitor))
		}
	}
}

// -----------------------------------------------------------------
// NOTIFICATIONS
// -----------------------------------------------------------------

// WithNotifications will set the notifications config
func WithNotifications(webhookEndpoint string) ClientOps {
	return func(c *clientOptions) {
		if len(webhookEndpoint) > 0 {
			c.notifications.webhookEndpoint = webhookEndpoint
		}
	}
}

// WithCustomNotifications will set a custom notifications interface
func WithCustomNotifications(customNotifications notifications.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if customNotifications != nil {
			c.notifications.ClientInterface = customNotifications
		}
	}
}

// WithMapiFeeQuotes will set usage of mapi fee quotes instead of default fees
func WithMapiFeeQuotes() ClientOps {
	return func(c *clientOptions) {
		c.chainstate.options = append(c.chainstate.options, chainstate.WithMapiFeeQuotes())
	}
}

// WithOverridenMAPIConfig will override mApi config with custom data(custom token, endpoints and etc.)
func WithOverridenMAPIConfig(miners []*minercraft.Miner) ClientOps {
	return func(c *clientOptions) {
		c.chainstate.options = append(c.chainstate.options, chainstate.WithOverridenMAPIConfig(miners))
	}
}

// WithMinercraft will set custom minercraft client
func WithMinercraft(minercraft minercraft.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if minercraft != nil {
			c.chainstate.options = append(c.chainstate.options, chainstate.WithMinercraft(minercraft))
		}
	}
}
