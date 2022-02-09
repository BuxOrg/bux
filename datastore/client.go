package datastore

import (
	"context"
	"strings"

	"github.com/newrelic/go-agent/v3/newrelic"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type (

	// Client is the datastore client (configuration)
	Client struct {
		options *clientOptions
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
		autoMigrate     bool             // Setting for Auto Migration of SQL tables
		db              *gorm.DB         // Database connection for Read-Only requests (can be same as Write)
		debug           bool             // Setting for global debugging
		engine          Engine           // Datastore engine (MySQL, PostgreSQL, SQLite)
		logger          logger.Interface // Custom logger interface
		migratedModels  []string         // List of models (types) that have been migrated
		migrateModels   []interface{}    // Models for migrations
		mongoDB         *mongo.Database  // Database connection for a MongoDB datastore
		mongoDBConfig   *MongoDBConfig   // Configuration for a MongoDB datastore
		newRelicEnabled bool             // If NewRelic is enabled (parent application)
		sqlConfigs      []*SQLConfig     // Configuration for a MySQL or PostgreSQL datastore
		sqLite          *SQLiteConfig    // Configuration for a SQLite datastore
		tablePrefix     string           // Model table prefix
	}
)

// NewClient creates a new client for all Datastore functionality
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

	// EMPTY! Engine was NOT set and will use the default (file based)
	if client.Engine().IsEmpty() {

		// Use default SQLite
		// Create a SQLite engine config
		opt := WithSQLite(&SQLiteConfig{
			CommonConfig: CommonConfig{
				Debug:       client.options.debug,
				TablePrefix: defaultTablePrefix,
			},
			DatabasePath: defaultSQLiteFileName,
			Shared:       defaultSQLiteSharing,
		})
		opt(client.options)
	}

	// Use NewRelic if it's enabled (use existing txn if found on ctx)
	ctx = client.options.getTxnCtx(ctx)

	// If NewRelic is enabled
	txn := newrelic.FromContext(ctx)
	if txn != nil {
		segment := txn.StartSegment("load_datastore")
		segment.AddAttribute("engine", client.Engine().String())
		defer segment.End()
	}

	// Use different datastore configurations
	var err error
	if client.Engine() == MySQL || client.Engine() == PostgreSQL {
		if client.options.db, err = openSQLDatabase(
			client.options.logger, client.options.sqlConfigs...,
		); err != nil {
			return nil, err
		}
	} else if client.Engine() == MongoDB {
		if client.options.mongoDB, err = openMongoDatabase(
			ctx, client.options.mongoDBConfig,
		); err != nil {
			return nil, err
		}
	} else { // SQLite
		if client.options.db, err = openSQLiteDatabase(
			client.options.logger, client.options.sqLite,
		); err != nil {
			return nil, err
		}
	}

	// Auto migrate
	if client.options.autoMigrate && len(client.options.migrateModels) > 0 {
		if err = client.AutoMigrateDatabase(ctx, client.options.migrateModels...); err != nil {
			return nil, err
		}
	}

	// Set logger if not set now
	if client.options.logger == nil {
		client.options.logger = newBasicLogger(client.IsDebug())
	}

	// Return the client
	return client, nil
}

// Close will terminate (close) the datastore and any open connections
func (c *Client) Close(ctx context.Context) error {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("close_datastore").End()
	}

	// Close Mongo
	if c.Engine() == MongoDB {
		if err := c.options.mongoDB.Client().Disconnect(ctx); err != nil {
			return err
		}
		c.options.mongoDB = nil
	} else { // All other SQL database(s)
		if err := closeSQLDatabase(c.options.db); err != nil {
			return err
		}
		c.options.db = nil
	}

	c.options.engine = Empty
	return nil
}

// IsDebug will return the debug flag (bool)
func (c *Client) IsDebug() bool {
	return c.options.debug
}

// Debug will set the debug flag
func (c *Client) Debug(on bool) {
	c.options.debug = on
}

// Engine will return the client's engine
func (c *Client) Engine() Engine {
	return c.options.engine
}

// IsNewRelicEnabled will return if new relic is enabled
func (c *Client) IsNewRelicEnabled() bool {
	return c.options.newRelicEnabled
}

// DebugLog will display verbose logs
func (c *Client) DebugLog(text string) {
	if c.options.debug && c.options.logger != nil {
		c.options.logger.Info(context.Background(), text)
	}
}

// HasMigratedModel will return if the model type has been migrated
func (c *Client) HasMigratedModel(modelType string) bool {
	for _, t := range c.options.migratedModels {
		if strings.EqualFold(t, modelType) {
			return true
		}
	}
	return false
}

// GetTableName will return the full table name for the given model name
func (c *Client) GetTableName(modelName string) string {
	if c.options.tablePrefix != "" {
		return c.options.tablePrefix + "_" + modelName
	}
	return modelName
}

// GetDatabaseName will return the full database name for the given model name
func (c *Client) GetDatabaseName() string {
	if c.Engine() == MySQL || c.Engine() == PostgreSQL {
		return c.options.sqlConfigs[0].Name
	} else if c.Engine() == MongoDB {
		return c.options.mongoDBConfig.DatabaseName
	}

	return ""
}
