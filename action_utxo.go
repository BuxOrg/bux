package bux

import (
	"context"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
)

// GetUtxos will get utxos based on an xPub
func (c *Client) GetUtxos(ctx context.Context, xPubID string, metadata *Metadata, conditions *map[string]interface{},
	queryParams *datastore.QueryParams) ([]*Utxo, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_utxos")

	// Get the utxos
	utxos, err := getUtxosByXpubID(
		ctx,
		xPubID,
		metadata,
		conditions,
		queryParams,
		c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

// GetUtxo will get a single utxo based on an xPub, the tx ID and the outputIndex
func (c *Client) GetUtxo(ctx context.Context, xPubKey, txID string, outputIndex uint32) (*Utxo, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_utxo")

	// Get the utxos
	utxo, err := getUtxo(
		ctx, txID, outputIndex, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}
	if utxo == nil {
		return nil, ErrMissingUtxo
	}

	// Check that the id matches
	if utxo.XpubID != utils.Hash(xPubKey) {
		return nil, ErrXpubIDMisMatch
	}

	return utxo, nil
}
