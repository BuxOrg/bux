package filters

import (
	"github.com/mrz1836/go-whatsonchain"
	"strings"

	"github.com/libsv/go-bt"
)

const PlanariaDTemplate = "006a223139694733575459537362796f7333754a373333794b347a45696f69314665734e55"
const PlanariaDTemplateAlternate = "6a223139694733575459537362796f7333754a373333794b347a45696f69314665734e55"

func PlanariaD(tx *whatsonchain.TxInfo) (*bt.Tx, error) {
	// Loop through all of the outputs and check for pubkeyhash output
	for _, out := range tx.Vout {
		// if any output contains a pubkeyhash output, include this tx in the filter
		if strings.HasPrefix(out.ScriptPubKey.Hex, PlanariaDTemplate) || strings.HasPrefix(out.ScriptPubKey.Hex, PlanariaDTemplateAlternate) {
			return bt.NewTxFromString(tx.Hex)
		}
	}
	return nil, nil
}
