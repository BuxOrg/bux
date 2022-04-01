package datastore

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// StorageService is the storage related methods
type StorageService interface {
	AutoMigrateDatabase(ctx context.Context, models ...interface{}) error
	Execute(query string) *gorm.DB
	GetModel(ctx context.Context, model interface{}, conditions map[string]interface{}, timeout time.Duration) error
	GetModels(ctx context.Context, models interface{}, conditions map[string]interface{}, pageSize, page int,
		orderByField, sortDirection string, fields *[]string, timeout time.Duration) error
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
	IsAutoMigrate() bool
	IsDebug() bool
	IsNewRelicEnabled() bool
}
