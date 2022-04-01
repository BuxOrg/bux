package chainstate

import (
	"fmt"
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
		fmt.Printf("Monitor: %s", item)
		m.Add(item)
	}

	return nil
}

/*func (p *Processor) FilterMempoolPublishEvent(e centrifuge.ServerPublishEvent) (*bt.Tx, error) {
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
*/
