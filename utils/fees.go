package utils

import (
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

// FeeUnit fee unit imported from go-bt/v2
type FeeUnit bt.FeeUnit

// IsLowerThan compare two fee units
func (f *FeeUnit) IsLowerThan(other *FeeUnit) bool {
	return float64(f.Satoshis)/float64(f.Bytes) < float64(other.Satoshis)/float64(other.Bytes)
}

// String returns the fee unit as a string
func (f *FeeUnit) String() string {
	return fmt.Sprintf("FeeUnit(%d satoshis / %d bytes)", f.Satoshis, f.Bytes)
}

// LowestFee get the lowest fee from a list of fee units AND a default value
func LowestFee(feeUnits []FeeUnit, defaultValue FeeUnit) FeeUnit {
	minFee := defaultValue
	for _, feeUnit := range feeUnits {
		if feeUnit.IsLowerThan(&minFee) {
			minFee = feeUnit
		}
	}
	return minFee
}

// GetInputSizeForType get an estimated size for the input based on the type
func GetInputSizeForType(inputType string) uint64 {
	switch inputType {
	case ScriptTypePubKeyHash:
		// 32 bytes txID
		// + 4 bytes vout index
		// + 1 byte script length
		// + 107 bytes script pub key
		// + 4 bytes nSequence
		return 148
	}

	return 500
}

// GetOutputSize get an estimated size for the output based on the type
func GetOutputSize(lockingScript string) uint64 {
	if lockingScript != "" {
		size, _ := hex.DecodeString(lockingScript)
		if size != nil {
			return uint64(len(size)) + 9 // 9 bytes = 8 bytes value, 1 byte length
		}
	}

	return 500
}
