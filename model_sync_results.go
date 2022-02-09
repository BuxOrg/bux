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
	Attempts    []*SyncAttempt `json:"attempts"`     // Each attempt
	LastMessage string         `json:"last_message"` // Last message (success or failure)
}

// SyncAttempt is the complete attempt to sync (multiple providers and strategies)
type SyncAttempt struct {
	Action        string    `json:"action"`         // type: broadcast, sync etc
	AttemptedAt   time.Time `json:"attempted_at"`   // Time it was attempted
	StatusMessage string    `json:"status_message"` // Success or failure message
	// Providers string `json:"providers"` // Provider used for attempt(s)
	// StatusCode & response info
	// Error message (if detected)
	// Miner or provider info
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
