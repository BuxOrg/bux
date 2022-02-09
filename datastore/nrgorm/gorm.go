package nrgorm

import (
	"fmt"
	"strings"

	"github.com/newrelic/go-agent/v3/newrelic"
	"gorm.io/gorm"
)

const (
	txnGormKey   = "newrelicTransaction"
	startTimeKey = "newrelicStartTime"
)

// SetTxnToGorm sets transaction to gorm settings, returns cloned DB
func SetTxnToGorm(txn *newrelic.Transaction, db *gorm.DB) *gorm.DB {
	if txn == nil {
		return db
	}
	return db.Set(txnGormKey, *txn) // todo: postgres wanted a val not pointer
}

// AddGormCallbacks adds callbacks to NewRelic, you should call SetTxnToGorm to make them work
func AddGormCallbacks(db *gorm.DB) {
	dialect := db.Dialector.Name()
	var product newrelic.DatastoreProduct
	switch dialect {
	case "postgres":
		product = newrelic.DatastorePostgres
	case "mysql":
		product = newrelic.DatastoreMySQL
	case "sqlite3":
		product = newrelic.DatastoreSQLite
	case "mssql":
		product = newrelic.DatastoreMSSQL
	default:
		return
	}
	c := newCallbacks(product)
	registerCallbacks(db, "transaction", c)
	registerCallbacks(db, "create", c)
	registerCallbacks(db, "query", c)
	registerCallbacks(db, "update", c)
	registerCallbacks(db, "delete", c)
	registerCallbacks(db, "row", c)
	registerCallbacks(db, "raw", c)
}

// callbacks is an internal type for registering all callbacks
type callbacks struct {
	product newrelic.DatastoreProduct
}

func newCallbacks(product newrelic.DatastoreProduct) *callbacks {
	return &callbacks{product}
}

func (c *callbacks) beforeCreate(tx *gorm.DB)   { c.before(tx) }
func (c *callbacks) afterCreate(tx *gorm.DB)    { c.after(tx, "INSERT") }
func (c *callbacks) beforeQuery(tx *gorm.DB)    { c.before(tx) }
func (c *callbacks) afterQuery(tx *gorm.DB)     { c.after(tx, "SELECT") }
func (c *callbacks) beforeUpdate(tx *gorm.DB)   { c.before(tx) }
func (c *callbacks) afterUpdate(tx *gorm.DB)    { c.after(tx, "UPDATE") }
func (c *callbacks) beforeDelete(tx *gorm.DB)   { c.before(tx) }
func (c *callbacks) afterDelete(tx *gorm.DB)    { c.after(tx, "DELETE") }
func (c *callbacks) beforeRowQuery(tx *gorm.DB) { c.before(tx) }
func (c *callbacks) afterRowQuery(tx *gorm.DB)  { c.after(tx, "") }

func (c *callbacks) before(tx *gorm.DB) {
	txn, ok := tx.Get(txnGormKey)
	if !ok {
		return
	}
	nrTxn := txn.(newrelic.Transaction)
	tx.Set(startTimeKey, nrTxn.StartSegment("before_hook").StartTime)
}

func (c *callbacks) after(tx *gorm.DB, operation string) {
	startTime, ok := tx.Get(startTimeKey)
	if !ok {
		return
	}
	if operation == "" {
		operation = strings.ToUpper(strings.Split(tx.Statement.SQL.String(), " ")[0])
	}
	segmentBuilder(
		startTime.(newrelic.SegmentStartTime),
		c.product,
		tx.Statement.SQL.String(),
		operation,
		tx.Statement.Table,
	).End()

	// gorm wraps insert&update into transaction automatically
	// add another segment for commit/rollback in such case
	if _, ok = tx.InstanceGet("gorm:started_transaction"); !ok {
		tx.Set(startTimeKey, nil)
		return
	}
	var txn interface{}
	if txn, ok = tx.Get(txnGormKey); !ok {
		return
	}
	nrTxn := txn.(newrelic.Transaction)
	tx.Set(startTimeKey, nrTxn.StartSegment("after_hook").StartTime)
}

func (c *callbacks) commitOrRollback(tx *gorm.DB) {
	startTime, ok := tx.Get(startTimeKey)
	if !ok || startTime == nil {
		return
	}

	segmentBuilder(
		startTime.(newrelic.SegmentStartTime),
		c.product,
		"",
		"COMMIT/ROLLBACK",
		tx.Statement.Table,
	).End()
}

func registerCallbacks(db *gorm.DB, name string, c *callbacks) {
	beforeName := fmt.Sprintf("newrelic:%v_before", name)
	afterName := fmt.Sprintf("newrelic:%v_after", name)
	gormCallbackName := fmt.Sprintf("gorm:%v", name)
	// gorm does some magic, if you pass CallbackProcessor here - nothing works
	switch name {
	case "create":
		_ = db.Callback().Create().Before(gormCallbackName).Register(beforeName, c.beforeCreate)
		_ = db.Callback().Create().After(gormCallbackName).Register(afterName, c.afterCreate)
		_ = db.Callback().Create().
			After("gorm:commit_or_rollback_transaction").
			Register(fmt.Sprintf("newrelic:commit_or_rollback_transaction_%v", name), c.commitOrRollback)
	case "query":
		_ = db.Callback().Query().Before(gormCallbackName).Register(beforeName, c.beforeQuery)
		_ = db.Callback().Query().After(gormCallbackName).Register(afterName, c.afterQuery)
	case "update":
		_ = db.Callback().Update().Before(gormCallbackName).Register(beforeName, c.beforeUpdate)
		_ = db.Callback().Update().After(gormCallbackName).Register(afterName, c.afterUpdate)
		_ = db.Callback().Update().
			After("gorm:commit_or_rollback_transaction").
			Register(fmt.Sprintf("newrelic:commit_or_rollback_transaction_%v", name), c.commitOrRollback)
	case "delete":
		_ = db.Callback().Delete().Before(gormCallbackName).Register(beforeName, c.beforeDelete)
		_ = db.Callback().Delete().After(gormCallbackName).Register(afterName, c.afterDelete)
		_ = db.Callback().Delete().
			After("gorm:commit_or_rollback_transaction").
			Register(fmt.Sprintf("newrelic:commit_or_rollback_transaction_%v", name), c.commitOrRollback)
	case "row":
		_ = db.Callback().Row().Before(gormCallbackName).Register(beforeName, c.beforeRowQuery)
		_ = db.Callback().Row().After(gormCallbackName).Register(afterName, c.afterRowQuery)
	case "raw":
		_ = db.Callback().Raw().Before(gormCallbackName).Register(beforeName, c.beforeRowQuery)
		_ = db.Callback().Raw().After(gormCallbackName).Register(afterName, c.afterRowQuery)
	}
}

type segment interface {
	End()
}

// create segment through function to be able to test it
var segmentBuilder = func(
	startTime newrelic.SegmentStartTime,
	product newrelic.DatastoreProduct,
	query string,
	operation string,
	collection string,
) segment {
	return &newrelic.DatastoreSegment{
		StartTime:          startTime,
		Product:            product,
		ParameterizedQuery: query,
		Operation:          operation,
		Collection:         collection,
	}
}
