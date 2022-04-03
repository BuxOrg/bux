package chainstate

import (
	"encoding/json"
	"github.com/BuxOrg/bux/chainstate/filters"
	"github.com/centrifugal/centrifuge-go"
	"github.com/libsv/go-bt/v2"
	"github.com/mrz1836/go-whatsonchain"
	boom "github.com/tylertreat/BoomFilters"
)

type BloomProcessor struct {
	filter *boom.StableBloomFilter
}

func NewBloomProcessor(maxCells uint, falsePositiveRate float64) *BloomProcessor {
	return &BloomProcessor{
		filter: boom.NewDefaultStableBloomFilter(maxCells, falsePositiveRate),
	}
}

type MonitorProcessor interface {
	Add(item string)
	Test(item string) bool
	Reload(items []string) error
	AddFilter(filterType TransactionType)
	FilterMempoolPublishEvent(event centrifuge.ServerPublishEvent) (*bt.Tx, error)
}

// Add a new item to the bloom filter
func (m *BloomProcessor) Add(item string) {
	m.filter.Add([]byte(item))
}

// Test checks whether the item is in the bloom filter
func (m *BloomProcessor) Test(item string) bool {
	return m.filter.Test([]byte(item))
}

// Reload the bloom filter from the DB
func (m *BloomProcessor) Reload(items []string) error {
	for _, item := range items {
		m.Add(item)
	}

	return nil
}

func (m *BloomProcessor) AddFilter(filterType TransactionType) {
	switch filterType {
	case Metanet:
		m.Add(filters.MetanetScriptTemplate)
	case PlanariaB:
		m.Add(filters.PlanariaDTemplate)
		m.Add(filters.PlanariaBTemplateAlternate)
	case RareCandyFrogCartel:
		m.Add(filters.RareCandyFrogCartelScriptTemplate)
	}
}

func (p *BloomProcessor) FilterMempoolPublishEvent(e centrifuge.ServerPublishEvent) (*bt.Tx, error) {
	transaction := whatsonchain.TxInfo{}
	err := json.Unmarshal(e.Data, &transaction)
	if err != nil {
		return nil, err
	}
	passes := p.Test(transaction.Hex)
	if !passes {
		return nil, nil
	}
	return bt.NewTxFromString(transaction.Hex)
}
