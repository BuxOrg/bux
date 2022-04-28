package bux

import (
	"context"

	"github.com/BuxOrg/bux/datastore"
)

// AdminStats are statistics about the bux server
type AdminStats struct {
	Balance            uint64
	Destinations       uint64
	Transactions       uint64
	Paymails           uint64
	Utxos              uint64
	XPubs              uint64
	TransactionsPerDay map[string]uint64
}

// AdminInterface is the bux admin interface comprised of all services available for admins
type AdminInterface interface {
	GetStats() (*AdminStats, error)
	GetPaymailAddresses(ctx context.Context, metadataConditions *Metadata, conditions *map[string]interface{},
		queryParams *datastore.QueryParams) ([]*PaymailAddress, error)
	GetPaymailAddressesCount(ctx context.Context, metadataConditions *Metadata,
		conditions *map[string]interface{}) (int64, error)
	GetXPubs(ctx context.Context, metadataConditions *Metadata,
		conditions *map[string]interface{}, queryParams *datastore.QueryParams) ([]*Xpub, error)
	GetXPubsCount(ctx context.Context, metadataConditions *Metadata,
		conditions *map[string]interface{}) (int64, error)
}

// GetStats get admin stats
func (c *Client) GetStats() (*AdminStats, error) {
	return nil, nil
}
