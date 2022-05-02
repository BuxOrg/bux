package bux

import (
	"context"
)

// GetStats will get stats for a admin dashboard
func (c *Client) GetStats(ctx context.Context, opts ...ModelOps) (stats *AdminStats, err error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "admin_stats")

	var destinationsCount int64
	var transactionsCount int64
	var paymailsCount int64
	var utxosCount int64
	var xpubsCount int64

	if destinationsCount, err = getDestinationsCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if transactionsCount, err = getTransactionsCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if paymailsCount, err = getPaymailAddressesCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if utxosCount, err = getUtxosCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if xpubsCount, err = getXPubsCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	stats = &AdminStats{
		Balance:            0,
		Destinations:       destinationsCount,
		Transactions:       transactionsCount,
		Paymails:           paymailsCount,
		Utxos:              utxosCount,
		XPubs:              xpubsCount,
		TransactionsPerDay: nil,
	}

	return stats, nil
}
