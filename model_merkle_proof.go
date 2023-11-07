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

func offsetPair(offset uint64) uint64 {
	if offset%2 == 0 {
		return offset + 1
	}
	return offset - 1
}

func parentOffset(offset uint64) uint64 {
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

// ToBUMP transform Merkle Proof to BUMP
func (m *MerkleProof) ToBUMP() BUMP {
	bump := BUMP{}

	height := len(m.Nodes)
	if height == 0 {
		return bump
	}

	path := make([][]BUMPLeaf, 0)
	txIDPath := make([]BUMPLeaf, 2)

	offset := m.Index
	pairOffset := offsetPair(offset)

	txIDPath1 := BUMPLeaf{
		Offset: offset,
		Hash:   m.TxOrID,
		TxID:   true,
	}
	txIDPath2 := BUMPLeaf{
		Offset: offsetPair(offset),
	}
	if m.Nodes[0] != "*" {
		txIDPath2.Hash = m.Nodes[0]
	} else {
		txIDPath2.Duplicate = true
	}

	if offset < pairOffset {
		txIDPath[0] = txIDPath1
		txIDPath[1] = txIDPath2
	} else {
		txIDPath[0] = txIDPath2
		txIDPath[1] = txIDPath1
	}

	path = append(path, txIDPath)
	for i := 1; i < height; i++ {
		p := make([]BUMPLeaf, 0)
		offset = parentOffset(offset)

		leaf := BUMPLeaf{Offset: offset}

		isDuplicate := m.Nodes[i] == "*"
		if !isDuplicate {
			leaf.Hash = m.Nodes[i]
		} else {
			leaf.Duplicate = true
		}

		p = append(p, leaf)
		path = append(path, p)
	}
	bump.Path = path
	return bump
}
