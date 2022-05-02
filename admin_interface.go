package bux

import (
	"context"

	"github.com/BuxOrg/bux/datastore"
)

// AdminStats are statistics about the bux server
type AdminStats struct {
	Balance            int64                  `json:"balance"`
	Destinations       int64                  `json:"destinations"`
	Transactions       int64                  `json:"transactions"`
	Paymails           int64                  `json:"paymails"`
	Utxos              int64                  `json:"utxos"`
	XPubs              int64                  `json:"xpubs"`
	TransactionsPerDay map[string]interface{} `json:"transactions_per_day"`
	UtxosPerType       map[string]interface{} `json:"utxos_per_type"`
}

// AdminInterface is the bux admin interface comprised of all services available for admins
type AdminInterface interface {
	GetStats(ctx context.Context, opts ...ModelOps) (*AdminStats, error)
	GetPaymailAddresses(ctx context.Context, metadataConditions *Metadata, conditions *map[string]interface{},
		queryParams *datastore.QueryParams, opts ...ModelOps) ([]*PaymailAddress, error)
	GetPaymailAddressesCount(ctx context.Context, metadataConditions *Metadata,
		conditions *map[string]interface{}, opts ...ModelOps) (int64, error)
	GetXPubs(ctx context.Context, metadataConditions *Metadata,
		conditions *map[string]interface{}, queryParams *datastore.QueryParams, opts ...ModelOps) ([]*Xpub, error)
	GetXPubsCount(ctx context.Context, metadataConditions *Metadata,
		conditions *map[string]interface{}, opts ...ModelOps) (int64, error)
}
