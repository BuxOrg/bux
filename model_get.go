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

	/*
		// todo: add cache support here for basic model lookups
	*/

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
	queryParams *datastore.QueryParams,
	timeout time.Duration,
) error {
	// Attempt to Get the model (by model fields & given conditions)
	return datastore.GetModels(ctx, models, conditions, queryParams, nil, timeout)
}
