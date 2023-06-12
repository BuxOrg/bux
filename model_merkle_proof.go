package bux

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/libsv/go-bc"
)

// MerkleProof type
type MerkleProof bc.MerkleProof

// Scan scan value into Json, implements sql.Scanner interface
func (m *MerkleProof) Scan(value interface{}) error {
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
func (m MerkleProof) Value() (driver.Value, error) {
	if reflect.DeepEqual(m, MerkleProof{}) {
		return nil, nil
	}
	marshal, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return string(marshal), nil
}
