package bux

import (
	"context"
	"time"

	"github.com/mrz1836/go-datastore"
	"github.com/pkg/errors"
)

// Save will save the model(s) into the Datastore
func Save(ctx context.Context, model ModelInterface) (err error) {

	// Check for a client
	c := model.Client()
	if c == nil {
		return ErrMissingClient
	}

	// Check for a datastore
	ds := c.Datastore()
	if ds == nil {
		return ErrDatastoreRequired
	}
	// Create new Datastore transaction
	// @siggi: we need this to be in a callback context for Mongo
	// NOTE: a DB error is not being returned from here
	return ds.NewTx(ctx, func(tx *datastore.Transaction) (err error) {

		// Fire the before hooks (parent model)
		if model.IsNew() {
			if err = model.BeforeCreating(ctx); err != nil {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					err = errors.Wrap(err, rollbackErr.Error())
				}
				return
			}
		} else {
			if err = model.BeforeUpdating(ctx); err != nil {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					err = errors.Wrap(err, rollbackErr.Error())
				}
				return
			}
		}

		// Set the record's timestamps
		model.SetRecordTime(model.IsNew())

		// Start the list of models to Save
		modelsToSave := append(make([]ModelInterface, 0), model)

		// Add any child models (fire before hooks)
		if children := model.ChildModels(); len(children) > 0 {
			for _, child := range children {
				if child.IsNew() {
					if err = child.BeforeCreating(ctx); err != nil {
						rollbackErr := tx.Rollback()
						if rollbackErr != nil {
							err = errors.Wrap(err, rollbackErr.Error())
						}
						return
					}
				} else {
					if err = child.BeforeUpdating(ctx); err != nil {
						rollbackErr := tx.Rollback()
						if rollbackErr != nil {
							err = errors.Wrap(err, rollbackErr.Error())
						}
						return
					}
				}

				// Set the record's timestamps
				child.SetRecordTime(child.IsNew())
			}

			// Add to list for saving
			modelsToSave = append(modelsToSave, children...)
		}

		// Logs for saving models
		model.Client().Logger().Debug().Msgf("saving %d models...", len(modelsToSave))

		// Save all models (or fail!)
		for index := range modelsToSave {
			modelsToSave[index].Client().Logger().Debug().
				Str("modelID", modelsToSave[index].GetID()).
				Msgf("starting to save model: %s", modelsToSave[index].Name())
			if err = modelsToSave[index].Client().Datastore().SaveModel(
				ctx, modelsToSave[index], tx, modelsToSave[index].IsNew(), false,
			); err != nil {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					err = errors.Wrap(err, rollbackErr.Error())
				}
				return
			}
		}

		// Commit all the model(s) if needed
		if tx.CanCommit() {
			model.Client().Logger().Debug().Msg("committing db transaction...")
			if err = tx.Commit(); err != nil {
				return
			}
		}

		// Fire after hooks (only on commit success)
		var afterErr error
		for index := range modelsToSave {
			if modelsToSave[index].IsNew() {
				modelsToSave[index].NotNew() // NOTE: calling it before this method... after created assumes it's been saved already
				afterErr = modelsToSave[index].AfterCreated(ctx)
			} else {
				afterErr = modelsToSave[index].AfterUpdated(ctx)
			}
			if afterErr != nil {
				if err == nil { // First error - set the error
					err = afterErr
				} else { // Got more than one error, wrap it!
					err = errors.Wrap(err, afterErr.Error())
				}
			}
		}

		return
	})
}

// saveToCache will save the model to the cache using the given key(s)
//
// ttl of 0 will cache forever
func saveToCache(ctx context.Context, keys []string, model ModelInterface, ttl time.Duration) error { //nolint:nolintlint,unparam // this does not matter
	// NOTE: this check is in place in-case a model does not load its parent Client()
	if model.Client() != nil {
		for _, key := range keys {
			if err := model.Client().Cachestore().SetModel(ctx, key, model, ttl); err != nil {
				return err
			}
		}
	} else {
		model.Client().Logger().Debug().
			Str("modelID", model.GetID()).
			Msg("ignoring saveToCache: client or cachestore is missing")
	}
	return nil
}
