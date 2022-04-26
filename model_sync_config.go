package bux

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// SyncConfig is the configuration used for syncing a transaction (on-chain)
type SyncConfig struct {
	Broadcast   bool `json:"broadcast" toml:"broadcast" yaml:"broadcast"`             // Transaction should be broadcasted
	PaymailP2P  bool `json:"paymail_p2p" toml:"paymail_p2p" yaml:"paymail_p2p"`       // Transaction will be sent to all related paymail providers if P2P is detected
	SyncOnChain bool `json:"sync_on_chain" toml:"sync_on_chain" yaml:"sync_on_chain"` // Transaction should be checked that it's on-chain
	// Miner       string `json:"miner" toml:"miner" yaml:"miner"`  // Use a specific miner
	// DelayToBroadcast time.Duration `json:"delay_to_broadcast" toml:"delay_to_broadcast" yaml:"delay_to_broadcast"` // Delay for broadcasting
	// UseQuote // Use a specific fee quote or policy
	// miners: []miner{name, token, feeQuote}
	// default: miner
	// failover: miner
	// keep tx updated until x blocks?
}

// Scan will scan the value into Struct, implements sql.Scanner interface
func (t *SyncConfig) Scan(value interface{}) error {
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
func (t SyncConfig) Value() (driver.Value, error) {
	marshal, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return string(marshal), nil
}
