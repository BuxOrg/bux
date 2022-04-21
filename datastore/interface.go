package datastore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// StorageService is the storage related methods
type StorageService interface {
	AutoMigrateDatabase(ctx context.Context, models ...interface{}) error
	CreateInBatches(ctx context.Context, models interface{}, batchSize int) error
	Execute(query string) *gorm.DB
	GetModel(ctx context.Context, model interface{}, conditions map[string]interface{}, timeout time.Duration) error
	GetModels(ctx context.Context, models interface{}, conditions map[string]interface{}, queryParams *QueryParams,
		fieldResults interface{}, timeout time.Duration) error
	HasMigratedModel(modelType string) bool
	IncrementModel(ctx context.Context, model interface{},
		fieldName string, increment int64) (newValue int64, err error)
	IndexExists(tableName, indexName string) (bool, error)
	IndexMetadata(tableName, field string) error
	NewTx(ctx context.Context, fn func(*Transaction) error) error
	Raw(query string) *gorm.DB
	SaveModel(ctx context.Context, model interface{}, tx *Transaction, newRecord, commitTx bool) error
}

// ClientInterface is the Datastore client interface
type ClientInterface interface {
	StorageService
	Close(ctx context.Context) error
	Debug(on bool)
	DebugLog(text string)
	Engine() Engine
	GetDatabaseName() string
	GetTableName(modelName string) string
	GetMongoCollection(collectionName string) *mongo.Collection
	GetMongoCollectionByTableName(tableName string) *mongo.Collection
	IsAutoMigrate() bool
	IsDebug() bool
	IsNewRelicEnabled() bool
}
