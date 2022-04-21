package bux

import (
	"context"
	"errors"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/logger"
)

// taskCleanupDraftTransactions will clean up all old expired draft transactions
func taskCleanupDraftTransactions(ctx context.Context, logClient logger.Interface, opts ...ModelOps) error {

	logClient.Info(ctx, "running cleanup draft transactions task...")

	// Construct an empty model
	var models []DraftTransaction
	conditions := map[string]interface{}{
		statusField: DraftStatusDraft,
		// todo: add DB condition for date "expires_at": map[string]interface{}{"$lte": time.Now()},
	}

	queryParams := &datastore.QueryParams{
		Page:          1,
		PageSize:      20,
		OrderByField:  idField,
		SortDirection: datastore.SortAsc,
	}

	// Get the records
	if err := getModels(
		ctx, NewBaseModel(ModelNameEmpty, opts...).Client().Datastore(),
		&models, conditions, queryParams, defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil
		}
		return err
	}

	// Loop and update
	var err error
	timeNow := time.Now().UTC()
	for index := range models {
		if timeNow.After(models[index].ExpiresAt) {
			models[index].enrich(ModelDraftTransaction, opts...)
			models[index].Status = DraftStatusExpired
			if err = models[index].Save(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// taskProcessIncomingTransactions will process any incoming transactions found
func taskProcessIncomingTransactions(ctx context.Context, logClient logger.Interface, opts ...ModelOps) error {

	logClient.Info(ctx, "running process incoming transaction(s) task...")

	err := processIncomingTransactions(ctx, 10, opts...)
	if err == nil || errors.Is(err, datastore.ErrNoResults) {
		return nil
	}
	return err
}

// taskBroadcastTransactions will broadcast any transactions
func taskBroadcastTransactions(ctx context.Context, logClient logger.Interface, opts ...ModelOps) error {

	logClient.Info(ctx, "running broadcast transaction(s) task...")

	err := processBroadcastTransactions(ctx, 10, opts...)
	if err == nil || errors.Is(err, datastore.ErrNoResults) {
		return nil
	}
	return err
}

// taskSyncTransactions will sync any transactions
func taskSyncTransactions(ctx context.Context, logClient logger.Interface, opts ...ModelOps) error {

	logClient.Info(ctx, "running sync transaction(s) task...")

	err := processSyncTransactions(ctx, 10, opts...)
	if err == nil || errors.Is(err, datastore.ErrNoResults) {
		return nil
	}
	return err
}
