package chainstate

import (
	"encoding/json"
	"github.com/BuxOrg/bux/chainstate/filters"

	"github.com/centrifugal/centrifuge-go"
	"github.com/libsv/go-bt"
	"github.com/mrz1836/go-whatsonchain"
)

type Processor struct {
	FilterType TransactionType
}

func NewProcessor(txType TransactionType) *Processor {
	return &Processor{FilterType: txType}
}

func (p *Processor) FilterMempoolPublishEvent(e centrifuge.ServerPublishEvent) (*bt.Tx, error) {
	transaction := whatsonchain.TxInfo{}
	err := json.Unmarshal(e.Data, &transaction)
	if err != nil {
		return nil, err
	}

	switch p.FilterType {
	case Metanet:
		return filters.Metanet(&transaction)
	case PubKeyHash:
		return filters.PubKeyHash(&transaction)
	case PlanariaB:
		return filters.PlanariaB(&transaction)
	case PlanariaD:
		return filters.PlanariaD(&transaction)
	case RareCandyFrogCartel:
		return filters.RareCandyFrogCartel(&transaction)
	}
	return nil, nil
}
