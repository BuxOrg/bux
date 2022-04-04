package chainstate

import (
	"encoding/json"
	"strings"

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
	for _, out := range transaction.Vout {
		passes := p.Test(out.ScriptPubKey.Hex)
		if passes {
			return bt.NewTxFromString(transaction.Hex)
		}
	}
	return nil, nil
}

// This processor just uses regex checks to see if a raw hex string exists in a tx
// This is bound to have some false positives but is somewhat performant when filter set is small
type RegexProcessor struct {
	filter []string
}

func NewRegexProcessor() *RegexProcessor {
	return &RegexProcessor{
		filter: []string{},
	}
}

func (p *RegexProcessor) Add(item string) {
	p.filter = append(p.filter, item)
}

func (p *RegexProcessor) Test(item string) bool {
	for _, f := range p.filter {
		if strings.Contains(item, f) {
			return true
		}
	}
	return false
}

func (p *RegexProcessor) Reload(items []string) error {
	for _, i := range items {
		p.Add(i)
	}
	return nil
}

func (p *RegexProcessor) FilterMempoolPublishEvent(e centrifuge.ServerPublishEvent) (*bt.Tx, error) {
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

func (m *RegexProcessor) AddFilter(filterType TransactionType) {
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
