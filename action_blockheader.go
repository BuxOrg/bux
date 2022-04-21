package bux

import (
	"context"
	"fmt"

	"github.com/libsv/go-bc"
)

// RecordBlockHeader will Save a block header into the Datastore
//
// hash is the hash of the block header
// bh is the block header data
// opts are model options and can include "metadata"
func (c *Client) RecordBlockHeader(ctx context.Context, hash string, height uint32, bh bc.BlockHeader,
	opts ...ModelOps) (*BlockHeader, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "record_block_header")

	// Create the model & set the default options (gives options from client->model)
	newOpts := c.DefaultModelOptions(append(opts, New())...)
	blockHeader := newBlockHeader(hash, height, bh, newOpts...)

	// Ensure that we have a transaction id (created from the txHex)
	id := blockHeader.GetID()
	if len(id) == 0 {
		return nil, ErrMissingBlockHeaderHash
	}

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, fmt.Sprintf(lockKeyRecordBlockHeader, id), c.Cachestore(),
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

// GetUnsyncedBlockHeaders get all unsynced block headers
func (c *Client) GetUnsyncedBlockHeaders(ctx context.Context) ([]*BlockHeader, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_unsynced_blockheaders")

	return getUnsyncedBlockHeaders(ctx, c.DefaultModelOptions()...)
}

// GetLastBlockHeader get last block header
func (c *Client) GetLastBlockHeader(ctx context.Context) (*BlockHeader, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_last_blockheader")

	return getLastBlockHeader(ctx, c.DefaultModelOptions()...)
}

// GetBlockHeaderByHeight get the block header by height
func (c *Client) GetBlockHeaderByHeight(ctx context.Context, height uint32) (*BlockHeader, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_blockheader_by_height")

	return getBlockHeaderByHeight(ctx, height, c.DefaultModelOptions()...)
}
