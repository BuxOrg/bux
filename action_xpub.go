package bux

import (
	"context"
	"sort"

	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt"
	"github.com/mrz1836/go-whatsonchain"
)

// NewXpub will parse the xPub and Save it into the Datastore
//
// xPubKey is the raw public xPub
// opts are options and can include "metadata"
func (c *Client) NewXpub(ctx context.Context, xPubKey string, opts ...ModelOps) (*Xpub, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_xpub")

	// Create the model & set the default options (gives options from client->model)
	xPub := newXpub(
		xPubKey, c.DefaultModelOptions(append(opts, New())...)...,
	)

	// Save the model
	if err := xPub.Save(ctx); err != nil {
		return nil, err
	}

	// Return the created model
	return xPub, nil
}

// GetXpub will get an existing xPub from the Datastore
//
// xPubKey is the raw public xPub
func (c *Client) GetXpub(ctx context.Context, xPubKey string) (*Xpub, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_xpub")

	// Validate the xPub
	if _, err := utils.ValidateXPub(xPubKey); err != nil {
		return nil, err
	}

	// Get the xPub (by key - converts to id)
	xPub, err := getXpub(
		ctx, xPubKey, // Pass the context and key everytime (for now)
		c.DefaultModelOptions()..., // Passing down the Datastore and client information into the model
	)
	if err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	// Return the model
	return xPub, nil
}

// GetXpubByID will get an existing xPub from the Datastore
func (c *Client) GetXpubByID(ctx context.Context, xPubID string) (*Xpub, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_xpub_by_id")

	// Get the xPub (by key - converts to id)
	xPub, err := getXpubByID(
		ctx, xPubID,
		c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	// Return the model
	return xPub, nil
}

// ImportXpub will import a given xPub and all related destinations and transactions
func (c *Client) ImportXpub(ctx context.Context, xPubKey string, depth uint32, opts ...ModelOps) (*ImportResults, error) {

	// Validate the xPub
	xPub, err := utils.ValidateXPub(xPubKey)
	if err != nil {
		return nil, err
	}

	// todo: add opts to each model for metadata?
	// opts

	// Start an accumulator
	results := &ImportResults{Key: xPub.String()}

	// Derive internal addresses until depth
	c.Logger().Info(ctx, "Deriving internal addresses...")
	addressList := whatsonchain.AddressList{}
	var destination *Destination
	for i := uint32(0); i < depth; i++ {
		// log.Printf("path m/1/%v", i)
		if destination, err = c.NewDestination(
			ctx, xPub.String(), utils.ChainInternal, utils.ScriptTypePubKeyHash, nil,
		); err != nil {
			return nil, err
		}
		addressList.Addresses = append(addressList.Addresses, destination.Address)
		results.InternalAddresses++
	}

	// Derive external addresses until gap limit
	c.Logger().Info(ctx, "Deriving external addresses...")
	for i := uint32(0); i < depth; i++ {
		// log.Printf("path m/0/%v", i)
		if destination, err = c.NewDestination(
			ctx, xPub.String(), utils.ChainExternal, utils.ScriptTypePubKeyHash, nil,
		); err != nil {
			return nil, err
		}
		addressList.Addresses = append(addressList.Addresses, destination.Address)
		results.ExternalAddresses++
	}

	// Set the WOC client
	woc := c.Chainstate().WhatsOnChain()

	// Get all transactions for those addresses
	var allTransactions []*whatsonchain.HistoryRecord
	if allTransactions, err = getTransactionsFromAddresses(
		ctx, woc, addressList,
	); err != nil {
		return nil, err
	}

	// Remove any duplicate transactions from all historical txs
	allTransactions = removeDuplicates(allTransactions)

	// Set all the hashes
	txHashes := whatsonchain.TxHashes{}
	for _, t := range allTransactions {
		txHashes.TxIDs = append(txHashes.TxIDs, t.TxHash)
		results.TransactionsFound++
	}

	// Run the bulk transaction processor
	var rawTxs []string
	var txInfos whatsonchain.TxList
	if txInfos, err = woc.BulkRawTransactionDataProcessor(
		ctx, &txHashes,
	); err != nil {
		return nil, err
	}

	// Loop and build from the inputs
	var tx *bt.Tx
	for i := 0; i < len(txInfos); i++ {
		if tx, err = bt.NewTxFromString(
			txInfos[i].Hex,
		); err != nil {
			return nil, err
		}
		var inputs []whatsonchain.VinInfo
		for _, in := range tx.Inputs {
			// todo: upgrade and use go-bt v2
			vin := whatsonchain.VinInfo{
				TxID: in.PreviousTxID,
			}
			inputs = append(inputs, vin)
		}
		txInfos[i].Vin = inputs
		rawTxs = append(rawTxs, txInfos[i].Hex)
	}

	// Sort all transactions by block height
	c.Logger().Info(ctx, "Sorting transactions to be recorded...")
	sort.Slice(txInfos, func(i, j int) bool {
		return txInfos[i].BlockHeight < txInfos[j].BlockHeight
	})

	// Sort transactions that are in the same block by previous tx
	for i := 0; i < len(txInfos); i++ {
		info := txInfos[i]
		bh := info.BlockHeight
		var sameBlockTxs []*whatsonchain.TxInfo
		sameBlockTxs = append(sameBlockTxs, info)

		// Loop through all remaining txs until block height is not the same
		for j := i + 1; j < len(txInfos); j++ {
			if txInfos[j].BlockHeight == bh {
				sameBlockTxs = append(sameBlockTxs, txInfos[j])
			} else {
				break
			}
		}
		if len(sameBlockTxs) == 1 {
			continue
		}

		// Sort transactions by whether previous txs are referenced in the inputs
		sort.Slice(sameBlockTxs, func(i, j int) bool {
			for _, in := range sameBlockTxs[i].Vin {
				if in.TxID == sameBlockTxs[j].Hash {
					return false
				}
			}
			return true
		})
		copy(txInfos[i:i+len(sameBlockTxs)], sameBlockTxs)
		i += len(sameBlockTxs) - 1
	}

	// Record transactions in bux
	for _, rawTx := range rawTxs {
		if _, err = c.RecordTransaction(
			ctx, xPubKey, rawTx, "",
		); err != nil {
			return nil, err
		}
		results.TransactionsImported++
	}

	return results, nil
}
