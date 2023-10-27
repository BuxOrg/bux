package bux

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bt/v2"
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
	pathMap := make(map[string]bt.VarInt, 2)
	offset := m.Index
	op := offsetPair(offset)
	pathMap[m.TxOrID] = bt.VarInt(offset)
	pathMap[m.Nodes[0]] = bt.VarInt(op)
	cmp[0] = pathMap
	for i := 1; i < height; i++ {
		path := make(map[string]bt.VarInt, 1)
		offset = parentOffset(offset)
		path[m.Nodes[i]] = bt.VarInt(offset)
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

func (m *MerkleProof) ToBUMP() BUMP {
	bump := BUMP{}

	height := len(m.Nodes)
	if height == 0 {
		return bump
	}

	path := make([][]BUMPNode, 0)
	txIdPath := make([]BUMPNode, 2)

	offset := m.Index
	pairOffset := offsetPair(offset)

	txIdPath1 := BUMPNode{
		Offset: offset,
		Hash:   m.TxOrID,
		TxId:   true,
	}
	txIdPath2 := BUMPNode{
		Offset: offsetPair(offset),
		Hash:   m.Nodes[0],
	}

	if offset < pairOffset {
		txIdPath[0] = txIdPath1
		txIdPath[1] = txIdPath2
	} else {
		txIdPath[0] = txIdPath2
		txIdPath[1] = txIdPath1
	}

	path = append(path, txIdPath)
	for i := 1; i < height; i++ {
		p := make([]BUMPNode, 0)
		offset = parentOffset(offset)
		p = append(p, BUMPNode{
			Offset: offset,
			Hash:   m.Nodes[i],
		})
		path = append(path, p)
	}
	bump.Path = path
	return bump
}
