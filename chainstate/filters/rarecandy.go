package filters

import (
	"github.com/mrz1836/go-whatsonchain"
	"strings"

	"github.com/libsv/go-bt"
)

const RareCandyFrogCartelScriptTemplate = "a914179b4c7a45646a509473df5a444b6e18b723bd148876"

func RareCandyFrogCartel(tx *whatsonchain.TxInfo) (*bt.Tx, error) {
	// Loop through all of the outputs and check for pubkeyhash output
	for _, out := range tx.Vout {
		// if any output contains a pubkeyhash output, include this tx in the filter
		if strings.HasPrefix(out.ScriptPubKey.Hex, RareCandyFrogCartelScriptTemplate) {
			return bt.NewTxFromString(tx.Hex)
		}
	}
	return nil, nil
}
