package bux

import (
	"context"
)

// GetStats will get stats for a admin dashboard
func (c *Client) GetStats(ctx context.Context, opts ...ModelOps) (stats *AdminStats, err error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "admin_stats")

	var (
		destinationsCount  int64
		transactionsCount  int64
		paymailsCount      int64
		utxosCount         int64
		xpubsCount         int64
		transactionsPerDay map[string]interface{}
		utxosPerType       map[string]interface{}
	)

	if destinationsCount, err = getDestinationsCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if transactionsCount, err = getTransactionsCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	conditions := map[string]interface{}{
		"deleted_at": nil,
	}
	if paymailsCount, err = getPaymailAddressesCount(ctx, nil, &conditions, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if utxosCount, err = getUtxosCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if xpubsCount, err = getXPubsCount(ctx, nil, nil, c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if transactionsPerDay, err = getTransactionsAggregate(ctx, nil, nil, "created_at", c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	if utxosPerType, err = getUtxosAggregate(ctx, nil, nil, "type", c.DefaultModelOptions(opts...)...); err != nil {
		return nil, err
	}

	stats = &AdminStats{
		Balance:            0,
		Destinations:       destinationsCount,
		Transactions:       transactionsCount,
		Paymails:           paymailsCount,
		Utxos:              utxosCount,
		XPubs:              xpubsCount,
		TransactionsPerDay: transactionsPerDay,
		UtxosPerType:       utxosPerType,
	}

	return stats, nil
}
