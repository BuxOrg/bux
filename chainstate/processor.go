package chainstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/centrifugal/centrifuge-go"
	"github.com/mrz1836/go-whatsonchain"
	boom "github.com/tylertreat/BoomFilters"
)

// BloomProcessor bloom filter processor
type BloomProcessor struct {
	debug             bool
	logger            Logger
	filters           map[string]*BloomProcessorFilter
	maxCells          uint
	falsePositiveRate float64
}

// BloomProcessorFilter struct
type BloomProcessorFilter struct {
	filter *boom.StableBloomFilter
	regex  *regexp.Regexp
}

// NewBloomProcessor initialize a new bloom processor
func NewBloomProcessor(maxCells uint, falsePositiveRate float64) *BloomProcessor {
	return &BloomProcessor{
		filters:           make(map[string]*BloomProcessorFilter),
		maxCells:          maxCells,
		falsePositiveRate: falsePositiveRate,
	}
}

// MonitorProcessor struct that defines interface to all filter processors
type MonitorProcessor interface {
	Debug(bool)
	IsDebug() bool
	SetLogger(logger Logger)
	Logger() Logger
	Add(regexString, item string) error
	Test(regexString string, item string) bool
	Reload(regexString string, items []string) error
	FilterMempoolPublishEvent(event centrifuge.ServerPublishEvent) (string, error)
	FilterMempoolTx(txHex string) (string, error)
	GetHash() string
}

// Debug set debugging
func (p *BloomProcessor) Debug(debug bool) {
	p.debug = debug
}

// IsDebug return whether debugging is on/off
func (p *BloomProcessor) IsDebug() bool {
	return p.debug
}

// SetLogger set the logger
func (p *BloomProcessor) SetLogger(logger Logger) {
	p.logger = logger
}

// Logger return the logger
func (p *BloomProcessor) Logger() Logger {
	return p.logger
}

// GetHash get the hash of the current filter
func (p *BloomProcessor) GetHash() string {
	h := sha256.New()
	for _, f := range p.filters {
		if _, err := f.filter.WriteTo(h); err != nil {
			return ""
		}
	}
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}

// Add a new item to the bloom filter
func (p *BloomProcessor) Add(regexString, item string) error {

	// check whether this regex filter already exists, otherwise initialize it
	if p.filters[regexString] == nil {
		r, err := regexp.Compile(regexString)
		if err != nil {
			return err
		}
		p.filters[regexString] = &BloomProcessorFilter{
			filter: boom.NewDefaultStableBloomFilter(p.maxCells, p.falsePositiveRate),
			regex:  r,
		}
	}
	p.filters[regexString].filter.Add([]byte(item))

	return nil
}

// Test checks whether the item is in the bloom filter
func (p *BloomProcessor) Test(regexString, item string) bool {
	if p.filters[regexString] == nil {
		return false
	}
	return p.filters[regexString].filter.Test([]byte(item))
}

// Reload the bloom filter from the DB
func (p *BloomProcessor) Reload(regexString string, items []string) (err error) {
	for _, item := range items {
		if err = p.Add(regexString, item); err != nil {
			return
		}
	}

	return
}

// FilterMempoolPublishEvent check whether a filter matches a mempool tx event
func (p *BloomProcessor) FilterMempoolPublishEvent(e centrifuge.ServerPublishEvent) (string, error) {
	transaction := whatsonchain.TxInfo{}
	_ = json.Unmarshal(e.Data, &transaction)

	for _, f := range p.filters {
		lockingScripts := f.regex.FindAllString(string(e.Data), -1)
		for _, ls := range lockingScripts {
			passes := f.filter.Test([]byte(ls))
			if passes {
				tx := whatsonchain.TxInfo{}
				err := json.Unmarshal(e.Data, &tx)
				if err != nil {
					return "", err
				}
				return tx.Hex, nil
			}
		}
	}

	return "", nil
}

// FilterMempoolTx check whether a filter matches a mempool tx event
func (p *BloomProcessor) FilterMempoolTx(txHex string) (string, error) {

	for _, f := range p.filters {
		lockingScripts := f.regex.FindAllString(txHex, -1)
		for _, ls := range lockingScripts {
			passes := f.filter.Test([]byte(ls))
			if passes {
				return txHex, nil
			}
		}
	}

	return "", nil
}

// RegexProcessor simple regex processor
// This processor just uses regex checks to see if a raw hex string exists in a tx
// This is bound to have some false positives but is somewhat performant when filter set is small
type RegexProcessor struct {
	debug  bool
	logger Logger
	filter []string
}

// NewRegexProcessor initialize a new regex processor
func NewRegexProcessor() *RegexProcessor {
	return &RegexProcessor{
		filter: []string{},
	}
}

// Debug set debugging
func (p *RegexProcessor) Debug(debug bool) {
	p.debug = debug
}

// IsDebug return whether debugging is on/off
func (p *RegexProcessor) IsDebug() bool {
	return p.debug
}

// SetLogger set the logger
func (p *RegexProcessor) SetLogger(logger Logger) {
	p.logger = logger
}

// Logger return the logger
func (p *RegexProcessor) Logger() Logger {
	return p.logger
}

// Add a new item to the processor
func (p *RegexProcessor) Add(regex string, _ string) error {
	p.filter = append(p.filter, regex)
	return nil
}

// Test checks whether the item matches an item in the processor
func (p *RegexProcessor) Test(_ string, item string) bool {
	for _, f := range p.filter {
		if strings.Contains(item, f) {
			return true
		}
	}
	return false
}

// Reload the items of the processor to match against
func (p *RegexProcessor) Reload(_ string, items []string) (err error) {
	for _, i := range items {
		if err = p.Add(i, ""); err != nil {
			return
		}
	}
	return
}

// FilterMempoolPublishEvent check whether a filter matches a mempool tx event
func (p *RegexProcessor) FilterMempoolPublishEvent(e centrifuge.ServerPublishEvent) (string, error) {
	transaction := whatsonchain.TxInfo{}
	err := json.Unmarshal(e.Data, &transaction)
	if err != nil {
		return "", err
	}
	passes := p.Test("", transaction.Hex)
	if !passes {
		return "", nil
	}
	return transaction.Hex, nil
}

// FilterMempoolTx filters mempool transaction
func (p *RegexProcessor) FilterMempoolTx(hex string) (string, error) {
	passes := p.Test("", hex)
	if passes {
		return hex, nil
	}
	return "", nil
}

// GetHash get the hash of the filter
func (p *RegexProcessor) GetHash() string {
	return ""
}
