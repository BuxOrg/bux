package utils

import (
	"encoding/hex"

	"github.com/libsv/go-bt/v2"
)

// FeeUnit fee unit imported from go-bt/v2
type FeeUnit bt.FeeUnit

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
