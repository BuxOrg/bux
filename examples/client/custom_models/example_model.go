package main

import (
	"context"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
)

// Example is an example model
type Example struct {
	bux.Model    `bson:",inline"` // Base bux model
	ID           string           `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique record id" bson:"_id"`                                       // Unique identifier
	ExampleField string           `json:"example_field" toml:"example_field" yaml:"example_field" gorm:"<-:create;type:varchar(64);comment:This is an example string field" bson:"example_field"` // Example string field
}

// ModelExample is an example model
const ModelExample = "example"
const tableExamples = "examples"

// NewExample create new example model
func NewExample(exampleString string, opts ...bux.ModelOps) *Example {
	id, _ := utils.RandomHex(32)

	// Standardize and sanitize!
	return &Example{
		Model:        *bux.NewBaseModel(ModelExample, opts...),
		ExampleField: exampleString,
		ID:           id,
	}
}

// GetModelName returns the model name
func (e *Example) GetModelName() string {
	return ModelExample
}

// GetModelTableName returns the model db table name
func (e *Example) GetModelTableName() string {
	return tableExamples
}

// Save the model
func (e *Example) Save(ctx context.Context) (err error) {
	return bux.Save(ctx, e)
}

// GetID will get the ID
func (e *Example) GetID() string {
	return e.ID
}

// BeforeCreating is called before the model is saved to the DB
func (e *Example) BeforeCreating(_ context.Context) (err error) {
	e.DebugLog("starting: " + e.Name() + " BeforeCreating hook...")

	// Do something here!

	e.DebugLog("end: " + e.Name() + " BeforeCreating hook")
	return
}

// Migrate model specific migration
func (e *Example) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableExamples), bux.ModelMetadata.String())
}
