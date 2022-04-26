package bux

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// SyncResults is the results from all sync attempts (broadcast or sync)
type SyncResults struct {
	LastMessage string        `json:"last_message"` // Last message (success or failure)
	Results     []*SyncResult `json:"results"`      // Each result of a sync task
}

// Sync actions for syncing transactions
const (
	syncActionBroadcast = "broadcast" // Broadcast a transaction into the mempool
	syncActionP2P       = "p2p"       // Notify all paymail providers associated to the transaction
	syncActionSync      = "sync"      // Get on-chain data about the transaction (IE: block hash, height, etc)
)

// SyncResult is the complete attempt/result to sync (multiple providers and strategies)
type SyncResult struct {
	Action        string    `json:"action"`             // type: broadcast, sync etc
	ExecutedAt    time.Time `json:"executed_at"`        // Time it was executed
	Provider      string    `json:"provider,omitempty"` // Provider used for attempt(s)
	StatusMessage string    `json:"status_message"`     // Success or failure message
}

// Scan will scan the value into Struct, implements sql.Scanner interface
func (t *SyncResults) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	xType := fmt.Sprintf("%T", value)
	var byteValue []byte
	if xType == ValueTypeString {
		byteValue = []byte(value.(string))
	} else {
		byteValue = value.([]byte)
	}
	if bytes.Equal(byteValue, []byte("")) || bytes.Equal(byteValue, []byte("\"\"")) {
		return nil
	}

	return json.Unmarshal(byteValue, &t)
}

// Value return json value, implement driver.Valuer interface
func (t SyncResults) Value() (driver.Value, error) {
	marshal, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return string(marshal), nil
}
