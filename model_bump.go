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

// BUMPPaths represents a slice of BUMPs (BSV Unified Merkle Paths)
type BUMPPaths []BUMP

// BUMP represents BUMP (BSV Unified Merkle Path) format
type BUMP struct {
	BlockHeight uint64       `json:"blockHeight,string"`
	Path        [][]BUMPNode `json:"path"`
	// private field for storing already used offsets to avoid duplicate nodes
	allNodes []map[uint64]bool
}

// BUMPNode represents each BUMP path element
type BUMPNode struct {
	Offset    uint64 `json:"offset,string"`
	Hash      string `json:"hash"`
	TxID      bool   `json:"txid,omitempty"`
	Duplicate bool   `json:"duplicate,omitempty"`
}

// CalculateMergedBUMP calculates Merged BUMP from a slice of Merkle Proofs
func CalculateMergedBUMP(mp []MerkleProof) (BUMP, error) {
	bump := BUMP{}

	if len(mp) == 0 || mp == nil {
		return bump, nil
	}

	height := len(mp[0].Nodes)
	if height > maxCmpHeight {
		return bump,
			fmt.Errorf("BUMP cannot be higher than %d", maxCmpHeight)
	}

	for _, m := range mp {
		if height != len(m.Nodes) {
			return bump,
				errors.New("Merged BUMP cannot be obtained from Merkle Proofs of different heights")
		}
	}

	bump.Path = make([][]BUMPNode, height)
	bump.allNodes = make([]map[uint64]bool, height)
	for i := range bump.allNodes {
		bump.allNodes[i] = make(map[uint64]bool, 0)
	}

	for _, m := range mp {
		bumpToAdd := m.ToBUMP()
		err := bump.add(bumpToAdd)
		if err != nil {
			return BUMP{}, err
		}
	}

	for _, p := range bump.Path {
		sort.Slice(p, func(i, j int) bool {
			return p[i].Offset < p[j].Offset
		})
	}

	return bump, nil
}

func (bump *BUMP) add(b BUMP) error {
	if len(bump.Path) != len(b.Path) {
		return errors.New("BUMPs with different heights cannot be merged")
	}

	for i := range b.Path {
		for _, v := range b.Path[i] {
			_, value := bump.allNodes[i][v.Offset]
			if !value {
				bump.Path[i] = append(bump.Path[i], v)
				bump.allNodes[i][v.Offset] = true
				continue
			}
			if i == 0 && value && v.TxID {
				for j := range bump.Path[i] {
					if bump.Path[i][j].Offset == v.Offset {
						bump.Path[i][j] = v
					}
				}
			}
		}
	}

	return nil
}

// Bytes returns BUMPPaths bytes
func (bumpPaths *BUMPPaths) Bytes() []byte {
	var buff bytes.Buffer

	for _, bump := range *bumpPaths {
		bytes, _ := hex.DecodeString(bump.Hex())
		buff.Write(bytes)
	}

	return buff.Bytes()
}

// Hex returns BUMP in hex format
func (bump *BUMP) Hex() string {
	return bump.bytesBuffer().String()
}

func (bump *BUMP) bytesBuffer() *bytes.Buffer {
	var buff bytes.Buffer
	buff.WriteString(hex.EncodeToString(bt.VarInt(bump.BlockHeight).Bytes()))

	height := len(bump.Path)
	buff.WriteString(leadingZeroInt(height))

	for i := 0; i < height; i++ {
		nodes := bump.Path[i]

		nLeafs := len(nodes)
		buff.WriteString(hex.EncodeToString(bt.VarInt(nLeafs).Bytes()))
		for _, n := range nodes {
			buff.WriteString(hex.EncodeToString(bt.VarInt(n.Offset).Bytes()))
			buff.WriteString(fmt.Sprintf("%02x", flags(n.TxID, n.Duplicate)))
			decodedHex, _ := hex.DecodeString(n.Hash)
			buff.WriteString(hex.EncodeToString(bt.ReverseBytes(decodedHex)))
		}
	}
	return &buff
}

// In case the offset or height is less than 10, they must be written with a leading zero
func leadingZeroInt(i int) string {
	return fmt.Sprintf("%02x", i)
}

func flags(txID, duplicate bool) byte {
	var (
		dataFlag      byte = 00
		duplicateFlag byte = 01
		txIDFlag      byte = 02
	)

	if duplicate {
		return duplicateFlag
	}
	if txID {
		return txIDFlag
	}
	return dataFlag
}

// Scan scan value into Json, implements sql.Scanner interface
func (bump *BUMP) Scan(value interface{}) error {
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

	return json.Unmarshal(byteValue, &bump)
}

// Value return json value, implement driver.Valuer interface
func (bump BUMP) Value() (driver.Value, error) {
	if reflect.DeepEqual(bump, BUMP{}) {
		return nil, nil
	}
	marshal, err := json.Marshal(bump)
	if err != nil {
		return nil, err
	}

	return string(marshal), nil
}
