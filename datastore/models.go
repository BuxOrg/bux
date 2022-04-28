package datastore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/BuxOrg/bux/datastore/nrgorm"
	"github.com/BuxOrg/bux/utils"
	"github.com/newrelic/go-agent/v3/newrelic"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// SaveModel will take care of creating or updating a model (primary key based) (abstracting the database)
//
// value is a pointer to the model, IE: &Transaction{}
func (c *Client) SaveModel(
	ctx context.Context,
	model interface{},
	tx *Transaction,
	newRecord, commitTx bool,
) error {

	// MongoDB (does not support transactions at this time)
	if c.Engine() == MongoDB {
		sessionContext := ctx //nolint:contextcheck // we need to overwrite the ctx for transaction support
		if tx.mongoTx != nil {
			// set the context to the session context -> mongo transaction
			sessionContext = *tx.mongoTx
		}
		if err := c.saveWithMongo(sessionContext, model, newRecord); err != nil {
			return err
		}
		return nil
	} else if !IsSQLEngine(c.Engine()) {
		return ErrUnsupportedEngine
	}

	// Set the NewRelic txn
	c.options.db = nrgorm.SetTxnToGorm(newrelic.FromContext(ctx), c.options.db)

	// Capture any panics
	defer func() {
		if r := recover(); r != nil {
			c.DebugLog(fmt.Sprintf("panic recovered: %v", r))
			_ = tx.Rollback()
		}
	}()
	if err := tx.sqlTx.Error; err != nil {
		return err
	}

	// Create vs Update
	if newRecord {
		if err := tx.sqlTx.Omit(clause.Associations).Create(model).Error; err != nil {
			_ = tx.Rollback()
			// todo add duplicate key check for MySQL, Postgres and SQLite
			return err
		}
	} else {
		if err := tx.sqlTx.Omit(clause.Associations).Save(model).Error; err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// Commit & check for errors
	if commitTx {
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	// Return the tx
	return nil
}

// IncrementModel will increment the given field atomically in the database and return the new value
func (c *Client) IncrementModel(
	ctx context.Context,
	model interface{},
	fieldName string,
	increment int64,
) (newValue int64, err error) {

	if c.Engine() == MongoDB {
		return c.incrementWithMongo(ctx, model, fieldName, increment)
	} else if !IsSQLEngine(c.Engine()) {
		return 0, ErrUnsupportedEngine
	}

	// Set the NewRelic txn
	c.options.db = nrgorm.SetTxnToGorm(newrelic.FromContext(ctx), c.options.db)

	// Create a new transaction
	if err = c.options.db.Transaction(func(tx *gorm.DB) error {

		// Get the id of the model
		id := utils.GetModelStringAttribute(model, "ID")
		if id == nil {
			return errors.New("model is missing an ID field")
		}

		// Get model if exist
		var result map[string]interface{}
		if err = tx.Model(&model).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&result).Error; err != nil {
			return err
		}

		if result == nil {
			newValue = increment
			return nil
		}

		// Increment Counter
		newValue = convertToInt64(result[fieldName]) + increment
		return tx.Model(&model).Where("id = ?", id).Update(fieldName, newValue).Error
	}); err != nil {
		return
	}

	return
}

// CreateInBatches create all the models given in batches
func (c *Client) CreateInBatches(
	ctx context.Context,
	models interface{},
	batchSize int,
) error {
	if c.Engine() == MongoDB {
		return c.CreateInBatchesMongo(ctx, models, batchSize)
	}

	tx := c.options.db.CreateInBatches(models, batchSize)
	return tx.Error
}

// convertToInt64 will convert an interface to an int64
func convertToInt64(i interface{}) int64 {
	switch v := i.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	}

	return i.(int64)
}

type gormWhere struct {
	tx *gorm.DB
}

func (g *gormWhere) Where(query interface{}, args ...interface{}) {
	g.tx.Where(query, args...)
}

func (g *gormWhere) getGormTx() *gorm.DB {
	return g.tx
}

// GetModel will get a model from the datastore
func (c *Client) GetModel(
	ctx context.Context,
	model interface{},
	conditions map[string]interface{},
	timeout time.Duration,
	forceWriteDB bool,
) error {

	// Switch on the datastore engines
	if c.Engine() == MongoDB { // Get using Mongo
		return c.getWithMongo(ctx, model, conditions, nil, nil)
	} else if !IsSQLEngine(c.Engine()) {
		return ErrUnsupportedEngine
	}

	// Set the NewRelic txn
	c.options.db = nrgorm.SetTxnToGorm(newrelic.FromContext(ctx), c.options.db)

	// Create a new context, and new db tx
	ctxDB, cancel := createCtx(ctx, c.options.db, timeout, c.IsDebug(), c.options.loggerDB)
	defer cancel()

	// Get the model data using a select
	// todo: optimize by specific fields
	var tx *gorm.DB
	if forceWriteDB { // Use the "write" database for this query
		tx = ctxDB.Clauses(dbresolver.Write).Select("*")
	} else { // Use a replica if found
		tx = ctxDB.Select("*")
	}

	// Add conditions
	if len(conditions) > 0 {
		gtx := gormWhere{tx: tx}
		return checkResult(BuxWhere(&gtx, conditions, c.Engine()).(*gorm.DB).Find(model))
	}

	return checkResult(tx.Find(model))
}

// GetModels will return a slice of models based on the given conditions
func (c *Client) GetModels(
	ctx context.Context,
	models interface{},
	conditions map[string]interface{},
	queryParams *QueryParams,
	fieldResults interface{},
	timeout time.Duration,
) error {

	if queryParams == nil {
		// init a new empty object for the default queryParams
		queryParams = &QueryParams{}
	}
	// Set default page size
	if queryParams.Page > 0 && queryParams.PageSize < 1 {
		queryParams.PageSize = defaultPageSize
	}

	// lower case the sort direction (asc / desc)
	queryParams.SortDirection = strings.ToLower(queryParams.SortDirection)

	// Switch on the datastore engines
	if c.Engine() == MongoDB { // Get using Mongo
		return c.getWithMongo(ctx, models, conditions, fieldResults, queryParams)
	} else if !IsSQLEngine(c.Engine()) {
		return ErrUnsupportedEngine
	}
	return c.find(ctx, models, conditions, queryParams, fieldResults, timeout)
}

// GetModelCount will return a count of the model matching conditions
func (c *Client) GetModelCount(
	ctx context.Context,
	model interface{},
	conditions map[string]interface{},
	timeout time.Duration,
) (int64, error) {

	// Switch on the datastore engines
	if c.Engine() == MongoDB {
		return c.countWithMongo(ctx, model, conditions)
	} else if !IsSQLEngine(c.Engine()) {
		return 0, ErrUnsupportedEngine
	}

	return c.count(ctx, model, conditions, timeout)
}

// find will get records and return
func (c *Client) find(ctx context.Context, result interface{}, conditions map[string]interface{},
	queryParams *QueryParams, fieldResults interface{}, timeout time.Duration) error {

	// Find the type
	if reflect.TypeOf(result).Elem().Kind() != reflect.Slice {
		return errors.New("field: result is not a slice, found: " + reflect.TypeOf(result).Kind().String())
	}

	// Set the NewRelic txn
	c.options.db = nrgorm.SetTxnToGorm(newrelic.FromContext(ctx), c.options.db)

	// Create a new context, and new db tx
	ctxDB, cancel := createCtx(ctx, c.options.db, timeout, c.IsDebug(), c.options.loggerDB)
	defer cancel()

	tx := ctxDB.Model(result)

	// Create the offset
	offset := (queryParams.Page - 1) * queryParams.PageSize

	// Use the limit and offset
	if queryParams.Page > 0 && queryParams.PageSize > 0 {
		tx = tx.Limit(queryParams.PageSize).Offset(offset)
	}

	// Use an order field/sort
	if len(queryParams.OrderByField) > 0 {
		tx = tx.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: queryParams.OrderByField,
			},
			Desc: strings.ToLower(queryParams.SortDirection) == SortDesc,
		})
	}

	// Check for errors or no records found
	if len(conditions) > 0 {
		gtx := gormWhere{tx: tx}
		return checkResult(BuxWhere(&gtx, conditions, c.Engine()).(*gorm.DB).Find(result))
	}

	// Skip the conditions
	if fieldResults != nil {
		return checkResult(tx.Find(fieldResults))
	}
	return checkResult(tx.Find(result))
}

// find will get records and return
func (c *Client) count(ctx context.Context, model interface{}, conditions map[string]interface{},
	timeout time.Duration) (int64, error) {

	// Find the type
	if reflect.TypeOf(model).Elem().Kind() != reflect.Slice {
		return 0, errors.New("field: result is not a slice, found: " + reflect.TypeOf(model).Kind().String())
	}

	// Set the NewRelic txn
	c.options.db = nrgorm.SetTxnToGorm(newrelic.FromContext(ctx), c.options.db)

	// Create a new context, and new db tx
	ctxDB, cancel := createCtx(ctx, c.options.db, timeout, c.IsDebug(), c.options.loggerDB)
	defer cancel()

	tx := ctxDB.Model(model)

	// Check for errors or no records found
	if len(conditions) > 0 {
		gtx := gormWhere{tx: tx}
		var count int64
		err := checkResult(BuxWhere(&gtx, conditions, c.Engine()).(*gorm.DB).Model(model).Count(&count))
		return count, err
	}
	var count int64
	err := checkResult(tx.Count(&count))

	return count, err
}

// Execute a SQL query
func (c *Client) Execute(query string) *gorm.DB {
	if IsSQLEngine(c.Engine()) {
		return c.options.db.Exec(query)
	}

	return nil
}

// Raw a raw SQL query
func (c *Client) Raw(query string) *gorm.DB {
	if IsSQLEngine(c.Engine()) {
		return c.options.db.Raw(query)
	}

	return nil
}

// checkResult will check for records or error
func checkResult(result *gorm.DB) error {
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrNoResults
		}
		return result.Error
	}

	// We should actually have some rows according to GORM
	if result.RowsAffected == 0 {
		return ErrNoResults
	}
	return nil
}

// createCtx will make a new DB context
func createCtx(ctx context.Context, db *gorm.DB, timeout time.Duration, debug bool,
	optionalLogger logger.Interface) (*gorm.DB, context.CancelFunc) {

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	return db.Session(getGormSessionConfig(db.PrepareStmt, debug, optionalLogger)).WithContext(ctx), cancel
}
