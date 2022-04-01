/*
Package chainstate is the on-chain data service abstraction layer
*/
package chainstate

import (
	"context"
	"time"
)

// MonitorBlockHeaders will start up a block headers monitor
func (c *Client) MonitorBlockHeaders(ctx context.Context) error {
	return nil
}

// Broadcast will attempt to broadcast a transaction
func (c *Client) Broadcast(ctx context.Context, id, txHex string, timeout time.Duration) error {

	// Basic validation
	if len(id) < 50 {
		return ErrInvalidTransactionID
	} else if len(txHex) <= 0 { // todo: validate the tx hex
		return ErrInvalidTransactionHex
	}

	// Broadcast!
	err := c.broadcast(ctx, id, txHex, timeout)
	if err != nil {
		return err
	}

	// Check if in mempool?

	return nil
}

// QueryTransaction will get the transaction info from all providers returning the "first" valid result
//
// Note: this is slow, but follows a specific order: mAPI -> WhatsOnChain -> MatterCloud -> NowNodes
func (c *Client) QueryTransaction(
	ctx context.Context, id string, requiredIn RequiredIn, timeout time.Duration,
) (*TransactionInfo, error) {

	// Basic validation
	if len(id) < 50 {
		return nil, ErrInvalidTransactionID
	} else if !c.validRequirement(requiredIn) {
		return nil, ErrInvalidRequirements
	}

	// Try all providers and return the "first" valid response
	info := c.query(ctx, id, requiredIn, timeout)
	if info == nil {
		return nil, ErrTransactionNotFound
	}
	return info, nil
}

// QueryTransactionFastest will get the transaction info from ALL provider(s) returning the "fastest" valid result
//
// Note: this is fast but could abuse each provider based on how excessive this method is used
func (c *Client) QueryTransactionFastest(
	ctx context.Context, id string, requiredIn RequiredIn, timeout time.Duration,
) (*TransactionInfo, error) {

	// Basic validation
	if len(id) < 50 {
		return nil, ErrInvalidTransactionID
	} else if !c.validRequirement(requiredIn) {
		return nil, ErrInvalidRequirements
	}

	// Try all providers and return the "fastest" valid response
	info := c.fastestQuery(ctx, id, requiredIn, timeout)
	if info == nil {
		return nil, ErrTransactionNotFound
	}
	return info, nil
}
