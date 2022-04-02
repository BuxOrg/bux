package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/notifications"
)

// AfterDeleted will fire after a successful delete in the Datastore
func (m *Model) AfterDeleted(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterDelete hook...")
	m.DebugLog("end: " + m.Name() + " AfterDelete hook")
	return nil
}

// BeforeUpdating will fire before updating a model in the Datastore
func (m *Model) BeforeUpdating(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " BeforeUpdate hook...")
	m.DebugLog("end: " + m.Name() + " BeforeUpdate hook")
	return nil
}

// Client will return the current client
func (m *Model) Client() ClientInterface {
	return m.client
}

// ChildModels will return any child models
func (m *Model) ChildModels() []ModelInterface {
	return nil
}

// DebugLog will display verbose logs
func (m *Model) DebugLog(text string) {
	c := m.Client()
	if c != nil && c.IsDebug() {
		c.Logger().Info(context.Background(), text)
	}
}

// enrich is run after getting a record from the database
func (m *Model) enrich(name ModelName, opts ...ModelOps) {
	// Set the name
	m.name = name

	// Overwrite defaults
	m.SetOptions(opts...)
}

// GetOptions will get the options that are set on that model
func (m *Model) GetOptions(isNewRecord bool) (opts []ModelOps) {

	// Client was set on the model
	if m.client != nil {
		opts = append(opts, WithClient(m.client))
	}

	// New record flag
	if isNewRecord {
		opts = append(opts, New())
	}

	return
}

// IsNew returns true if the model is (or was) a new record
func (m *Model) IsNew() bool {
	return m.newRecord
}

// Name will get the collection name (model)
func (m *Model) Name() string {
	return m.name.String()
}

// New will set the record to new
func (m *Model) New() {
	m.newRecord = true
}

// NotNew sets newRecord to false
func (m *Model) NotNew() {
	m.newRecord = false
}

// RawXpub returns the rawXpubKey
func (m *Model) RawXpub() string {
	return m.rawXpubKey
}

// SetRecordTime will set the record timestamps (created is true for a new record)
func (m *Model) SetRecordTime(created bool) {
	if created {
		m.CreatedAt = time.Now().UTC()
	} else {
		m.UpdatedAt = time.Now().UTC()
	}
}

// UpdateMetadata will update the metadata on the model
// any key set to nil will be removed, other keys updated or added
func (m *Model) UpdateMetadata(metadata Metadata) {
	if m.Metadata == nil {
		m.Metadata = make(Metadata)
	}

	for key, value := range metadata {
		if value == nil {
			delete(m.Metadata, key)
		} else {
			m.Metadata[key] = value
		}
	}
}

// SetOptions will set the options on the model
func (m *Model) SetOptions(opts ...ModelOps) {
	for _, opt := range opts {
		opt(m)
	}
}

// Display filter the model for display
func (m *Model) Display() interface{} {
	return m
}

// RegisterTasks will register the model specific tasks on client initialization
func (m *Model) RegisterTasks() error {
	return nil
}

// AfterUpdated will fire after a successful update into the Datastore
func (m *Model) AfterUpdated(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterUpdated hook...")
	m.DebugLog("end: " + m.Name() + " AfterUpdated hook")
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *Model) AfterCreated(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")
	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

// Notify about an event on the model
// it's a bit weird to call this on the model, with the model and id as parameters, but this seems to be the
// easiest way to refactor away from the models themselves, with all the needed variables available.
// We cannot access client.Notifications() on the model, so need the m *Model
// We cannot access ID on the model and need id, mainly for error handling and reporting what went wrong :-/
func (m *Model) Notify(event notifications.EventType, model interface{}, id string) {
	// run the notifications in a separate goroutine since there could be significant network delay
	// communicating with a notification provider
	go func() {
		if n := m.client.Notifications(); n != nil {
			ctx := context.Background()
			if err := n.Notify(ctx, event, model, id); err != nil {
				m.Client().Logger().Error(context.Background(), "failed notifying about "+string(event)+" on "+id+": "+err.Error())
			}
		}
	}()
}
