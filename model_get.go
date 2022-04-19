package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/datastore"
)

// Get will retrieve a model from the Cachestore or Datastore using the provided conditions
//
// use bypassCache to skip checking the Cachestore for the record
func Get(
	ctx context.Context,
	model ModelInterface,
	conditions map[string]interface{},
	_ bool,
	timeout time.Duration,
) error {

	if timeout == 0 {
		timeout = defaultDatabaseReadTimeout
	}

	// Only use the cache if we are not bypassing...
	// if !bypassCache {

	// Do we have cache enabled on the model?
	// c := model.Client().Cachestore()
	// if c != nil {

	// todo: this does not work correctly, needs to use certain conditions
	// IE: if we are looking up an xPub by ID, check cache using that condition
	/*
		// Get the model from the cache
		if err := c.GetModel(
			ctx, fmt.Sprintf("%s-id-%s", model.GetModelName(), model.GetID()), model,
		); err != nil {

			// Only a REAL error will halt this request
			if !errors.Is(err, Cachestore.ErrKeyNotFound) {
				return err
			}

			// @mrz: This will continue gracefully if the record is not found

		} else { // No error means we got a result back!
			return nil
		}
	*/
	// }
	// }

	// Attempt to Get the model (by model fields & given conditions)
	return model.Client().Datastore().GetModel(ctx, model, conditions, timeout)
}

// getModels will retrieve model(s) from the Cachestore or Datastore using the provided conditions
//
// use bypassCache to skip checking the Cachestore for the record
func getModels(
	ctx context.Context,
	datastore datastore.ClientInterface,
	models interface{},
	conditions map[string]interface{},
	pageSize, page int,
	orderByField, sortDirection string,
	timeout time.Duration,
) error {
	// Attempt to Get the model (by model fields & given conditions)
	return datastore.GetModels(ctx, models, conditions, pageSize, page, orderByField, sortDirection, nil, timeout)
}
