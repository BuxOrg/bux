package bux

import (
	"context"
	"fmt"
)

// IncrementField will increment the given field atomically in the datastore
func IncrementField(ctx context.Context, model ModelInterface, fieldName string,
	increment int64) (int64, error) {

	// Debug log
	model.DebugLog(fmt.Sprintf("increment model %s field ... %s %d", model.Name(), fieldName, increment))

	// Increment
	return model.Client().Datastore().IncrementModel(ctx, model, fieldName, increment)
}
