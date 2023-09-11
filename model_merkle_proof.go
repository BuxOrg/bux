package bux

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/libsv/go-bc"
)

// MerkleProof represents Merkle Proof type
type MerkleProof bc.MerkleProof

// ToCompoundMerklePath transform Merkle Proof to Compound Merkle Path
func (m MerkleProof) ToCompoundMerklePath() CompoundMerklePath {
	height := len(m.Nodes)
	if height == 0 {
		return nil
	}
	cmp := make(CompoundMerklePath, height)
	pathMap := make(map[string]uint64, 2)
	offset := m.Index
	pathMap[m.TxOrID] = offset
	pathMap[m.Nodes[0]] = offsetPair(offset)
	cmp[0] = pathMap
	for i := 1; i < height; i++ {
		path := make(map[string]uint64, 1)
		offset = parrentOffset(offset)
		path[m.Nodes[i]] = offset
		cmp[i] = path
	}
	return cmp
}

func offsetPair(offset uint64) uint64 {
	if offset%2 == 0 {
		return offset + 1
	}
	return offset - 1
}

func parrentOffset(offset uint64) uint64 {
	return offsetPair(offset / 2)
}

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
