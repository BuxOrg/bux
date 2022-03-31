package filters

import (
	"strings"

	"github.com/libsv/go-bt"
	"github.com/mrz1836/go-whatsonchain"
)

const PlanariaBTemplate = "006a2231394878696756345179427633744870515663554551797131707a5a56646f417574"
const PlanariaBTemplateAlternate = "6a2231394878696756345179427633744870515663554551797131707a5a56646f417574"

func PlanariaB(tx *whatsonchain.TxInfo) (*bt.Tx, error) {
	// Loop through all of the outputs and check for pubkeyhash output
	for _, out := range tx.Vout {
		// if any output contains a pubkeyhash output, include this tx in the filter
		if strings.HasPrefix(out.ScriptPubKey.Hex, PlanariaBTemplate) || strings.HasPrefix(out.ScriptPubKey.Hex, PlanariaBTemplateAlternate) {
			return bt.NewTxFromString(tx.Hex)
		}
	}
	return nil, nil
}
