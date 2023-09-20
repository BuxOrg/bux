package bux

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/libsv/go-bt/v2"
)

// CompoundMerklePath represents Compound Merkle Path type
type CompoundMerklePath []map[string]bt.VarInt

// CMPSlice represents slice of Compound Merkle Pathes
// There must be several CMPs in case if utxos from different blocks is used in tx
type CMPSlice []CompoundMerklePath

type nodeOffset struct {
	node   string
	offset bt.VarInt
}

// Hex returns CMP in hex format
func (cmp *CompoundMerklePath) Hex() string {
	return cmp.bytesBuffer().String()
}

// Bytes returns CMP bytes
func (cmp *CompoundMerklePath) Bytes() []byte {
	return cmp.bytesBuffer().Bytes()
}

// Bytes returns CMPSlice bytes
func (cmps *CMPSlice) Bytes() []byte {
	var buff bytes.Buffer

	for _, cmp := range *cmps {
		buff.Write(cmp.Bytes())
	}

	return buff.Bytes()
}

func (cmp *CompoundMerklePath) bytesBuffer() *bytes.Buffer {
	var buff bytes.Buffer
	buff.WriteString(leadingZeroInt(len(*cmp) - 1))

	for _, m := range *cmp {
		leafs := len(m)
		buff.WriteString(hex.EncodeToString(bt.VarInt(leafs).Bytes()))
		sortedNodes := sortByOffset(m)
		for _, n := range sortedNodes {
			buff.WriteString(hex.EncodeToString(n.offset.Bytes()))
			buff.WriteString(n.node)
		}
	}
	return &buff
}

func sortByOffset(m map[string]bt.VarInt) []nodeOffset {
	n := make([]nodeOffset, 0)
	for node, offset := range m {
		n = append(n, nodeOffset{node, offset})
	}
	sort.Slice(n, func(i, j int) bool {
		return n[i].offset < n[j].offset
	})
	return n
}

// CalculateCompoundMerklePath calculates CMP from a slice of Merkle Proofs
func CalculateCompoundMerklePath(mp []MerkleProof) (CompoundMerklePath, error) {
	if len(mp) == 0 || mp == nil {
		return CompoundMerklePath{}, nil
	}
	height := len(mp[0].Nodes)
	for _, m := range mp {
		if height != len(m.Nodes) {
			return nil,
				errors.New("Compound Merkle Path cannot be obtained from Merkle Proofs of different heights")
		}
	}
	cmp := make(CompoundMerklePath, height)
	for _, m := range mp {
		cmpToAdd := m.ToCompoundMerklePath()
		err := cmp.add(cmpToAdd)
		if err != nil {
			return CompoundMerklePath{}, err
		}
	}
	return cmp, nil
}

// In case the offset or height is less than 10, they must be written with a leading zero
func leadingZeroInt(i int) string {
	return fmt.Sprintf("%02d", i)
}

func (cmp *CompoundMerklePath) add(c CompoundMerklePath) error {
	if len(*cmp) != len(c) {
		return errors.New("Compound Merkle Path with different height cannot be added")
	}
	for i := range c {
		for k, v := range c[i] {
			if (*cmp)[i] == nil {
				(*cmp)[i] = c[i]
				break
			}
			(*cmp)[i][k] = v
		}
	}
	return nil
}

// Scan scan value into Json, implements sql.Scanner interface
func (cmps *CMPSlice) Scan(value interface{}) error {
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

	return json.Unmarshal(byteValue, &cmps)
}

// Value return json value, implement driver.Valuer interface
func (cmps CMPSlice) Value() (driver.Value, error) {
	if reflect.DeepEqual(cmps, CMPSlice{}) {
		return nil, nil
	}
	marshal, err := json.Marshal(cmps)
	if err != nil {
		return nil, err
	}

	return string(marshal), nil
}
