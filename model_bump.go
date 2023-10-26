package bux

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// BUMP represents BUMP format
type BUMP struct {
	BlockHeight uint64          `json:"blockHeight,string"`
	Path        [][]BUMPLeaf `json:"path"`
}

// BUMPPathElement represents each BUMP path element
type BUMPLeaf struct {
	Offset    uint64 `json:"offset,string"`
	Hash      string `json:"hash,omitempty"`
	TxId      bool   `json:"txid,omitempty"`
	Duplicate bool   `json:"duplicate,omitempty"`
}

// Scan scan value into Json, implements sql.Scanner interface
func (m *BUMP) Scan(value interface{}) error {
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

	return json.Unmarshal(byteValue, &m)
}

// Value return json value, implement driver.Valuer interface
func (m BUMP) Value() (driver.Value, error) {
	if reflect.DeepEqual(m, BUMP{}) {
		return nil, nil
	}
	marshal, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return string(marshal), nil
}
