package chainstate

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/BuxOrg/bux/utils"
	"github.com/centrifugal/centrifuge-go"
	"github.com/mrz1836/go-whatsonchain"
	boom "github.com/tylertreat/BoomFilters"
)

// BloomProcessor bloom filter processor
type BloomProcessor struct {
	filter         *boom.StableBloomFilter
	filterType     string
	regex          *regexp.Regexp
	regexCheckOnly bool
}

// NewBloomProcessor initalize a new bloom processor
func NewBloomProcessor(maxCells uint, falsePositiveRate float64, filterType string, regexCheckOnly bool) *BloomProcessor {
	regex := utils.GetDestinationTypeRegex(filterType)
	if regex == nil {
		regex = utils.GetDestinationTypeRegex(utils.ScriptTypePubKeyHash)
	}

	return &BloomProcessor{
		filter:         boom.NewDefaultStableBloomFilter(maxCells, falsePositiveRate),
		filterType:     filterType,
		regex:          regex,
		regexCheckOnly: regexCheckOnly,
	}
}

// MonitorProcessor struct that defines interface to all filter processors
type MonitorProcessor interface {
	Add(item string)
	Test(item string) bool
	Reload(items []string) error
	FilterMempoolPublishEvent(event centrifuge.ServerPublishEvent) (string, []string, error)
}

// Add a new item to the bloom filter
func (m *BloomProcessor) Add(item string) {
	m.filter.Add([]byte(item))
}

// Test checks whether the item is in the bloom filter
func (m *BloomProcessor) Test(item string) bool {
	if m.regexCheckOnly {
		return true
	}
	return m.filter.Test([]byte(item))
}

// Reload the bloom filter from the DB
func (m *BloomProcessor) Reload(items []string) error {
	for _, item := range items {
		m.Add(item)
	}

	return nil
}

// FilterMempoolPublishEvent check whether a filter matches a mempool tx event
func (p *BloomProcessor) FilterMempoolPublishEvent(e centrifuge.ServerPublishEvent) (string, []string, error) {

	lockingScripts := p.regex.FindAllString(string(e.Data), -1)
	for _, ls := range lockingScripts {
		passes := p.Test(ls)
		if passes {
			transaction := whatsonchain.TxInfo{}
			err := json.Unmarshal(e.Data, &transaction)
			if err != nil {
				return "", []string{}, err
			}
			return transaction.Hex, lockingScripts, nil
		}
	}

	return "", []string{}, nil
}

// RegexProcessor simple regex processor
// This processor just uses regex checks to see if a raw hex string exists in a tx
// This is bound to have some false positives but is somewhat performant when filter set is small
type RegexProcessor struct {
	filter []string
}

// NewRegexProcessor initialize a new regex processor
func NewRegexProcessor() *RegexProcessor {
	return &RegexProcessor{
		filter: []string{},
	}
}

// Add a new item to the processor
func (p *RegexProcessor) Add(item string) {
	p.filter = append(p.filter, item)
}

// Test checks whether the item matches an item in the processor
func (p *RegexProcessor) Test(item string) bool {
	for _, f := range p.filter {
		if strings.Contains(item, f) {
			return true
		}
	}
	return false
}

// Reload the items of the processor to match against
func (p *RegexProcessor) Reload(items []string) error {
	for _, i := range items {
		p.Add(i)
	}
	return nil
}

// FilterMempoolPublishEvent check whether a filter matches a mempool tx event
func (p *RegexProcessor) FilterMempoolPublishEvent(e centrifuge.ServerPublishEvent) (string, error) {
	transaction := whatsonchain.TxInfo{}
	err := json.Unmarshal(e.Data, &transaction)
	if err != nil {
		return "", err
	}
	passes := p.Test(transaction.Hex)
	if !passes {
		return "", nil
	}
	return transaction.Hex, nil
}
