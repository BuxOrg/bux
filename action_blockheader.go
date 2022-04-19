package bux

import (
	"context"

	"github.com/libsv/go-bc"
)

// RecordBlockHeader will Save a block header into the Datastore
//
// hash is the hash of the block header
// bh is the block header data
// opts are model options and can include "metadata"
func (c *Client) RecordBlockHeader(ctx context.Context, hash string, bh bc.BlockHeader,
	opts ...ModelOps) (*BlockHeader, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "record_blockheader")

	// Create the model & set the default options (gives options from client->model)
	newOpts := c.DefaultModelOptions(append(opts, New())...)
	blockHeader := newBlockHeader(hash, bh, newOpts...)

	// Ensure that we have a transaction id (created from the txHex)
	id := blockHeader.GetID()
	if len(id) == 0 {
		return nil, ErrMissingBlockHeaderHash
	}

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, "action-record-block-header-"+id, c.Cachestore(),
	)
	defer unlock()
	if err != nil {
		return nil, err
	}

	// Process & save the transaction model
	if err = blockHeader.Save(ctx); err != nil {
		return nil, err
	}

	// Return the response
	return blockHeader, nil
}
