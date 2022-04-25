package chainstate

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/mrz1836/go-mattercloud"
	"github.com/mrz1836/go-nownodes"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/tonicpow/go-minercraft"
)

// query will try ALL providers in order and return the first "valid" response based on requirements
func (c *Client) query(ctx context.Context, id string, requiredIn RequiredIn,
	timeout time.Duration) *TransactionInfo {

	// Create a context (to cancel or timeout)
	ctxWithCancel, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// First: try all mAPI miners (Only supported on main and test right now)
	if c.Network() == MainNet || c.Network() == TestNet {
		for index := range c.options.config.mAPI.queryMiners {
			if c.options.config.mAPI.queryMiners[index] != nil {
				if res, err := queryMAPI(
					ctxWithCancel, c, c.Minercraft(), c.options.config.mAPI.queryMiners[index].Miner, id,
				); err == nil && checkRequirement(requiredIn, id, res) {
					return res
				}
			}
		}
	}

	// Next: try WhatsOnChain
	if resp, err := queryWhatsOnChain(
		ctxWithCancel, c, c.WhatsOnChain(), id,
	); err == nil && checkRequirement(requiredIn, id, resp) {
		return resp
	}

	// Next: try MatterCloud
	if resp, err := queryMatterCloud(
		ctxWithCancel, c, c.MatterCloud(), id,
	); err == nil && checkRequirement(requiredIn, id, resp) {
		return resp
	}

	// Next: try NowNodes (if loaded)
	nn := c.NowNodes()
	if nn != nil && c.Network() == MainNet {
		if resp, err := queryNowNodes(
			ctxWithCancel, c, nn, id,
		); err == nil && checkRequirement(requiredIn, id, resp) {
			return resp
		}
	}

	// No transaction information found
	return nil
}

// fastestQuery will try ALL providers on once and return the fastest "valid" response based on requirements
func (c *Client) fastestQuery(ctx context.Context, id string, requiredIn RequiredIn,
	timeout time.Duration) *TransactionInfo {

	// The channel for the internal results
	resultsChannel := make(
		chan *TransactionInfo,
		len(c.options.config.mAPI.queryMiners)+2,
	) // All miners & WhatsOnChain & MatterCloud

	// Create a context (to cancel or timeout)
	ctxWithCancel, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Loop each miner (break into a Go routine for each query)
	var wg sync.WaitGroup
	if c.Network() == MainNet || c.Network() == TestNet {
		for index := range c.options.config.mAPI.queryMiners {
			wg.Add(1)
			go func(
				ctx context.Context, client *Client,
				wg *sync.WaitGroup, miner *minercraft.Miner,
				id string, requiredIn RequiredIn,
			) {
				defer wg.Done()
				if res, err := queryMAPI(
					ctx, client, client.Minercraft(), miner, id,
				); err == nil && checkRequirement(requiredIn, id, res) {
					resultsChannel <- res
				}
			}(ctxWithCancel, c, &wg, c.options.config.mAPI.queryMiners[index].Miner, id, requiredIn)
		}
	}

	// Backup: WhatsOnChain
	wg.Add(1)
	go func(ctx context.Context, client *Client, id string, requiredIn RequiredIn) {
		defer wg.Done()
		if resp, err := queryWhatsOnChain(
			ctx, client, client.WhatsOnChain(), id,
		); err == nil && checkRequirement(requiredIn, id, resp) {
			resultsChannel <- resp
		}
	}(ctxWithCancel, c, id, requiredIn)

	// Backup: MatterCloud
	wg.Add(1)
	go func(ctx context.Context, client *Client, id string, requiredIn RequiredIn) {
		defer wg.Done()
		if resp, err := queryMatterCloud(
			ctx, client, client.MatterCloud(), id,
		); err == nil && checkRequirement(requiredIn, id, resp) {
			resultsChannel <- resp
		}
	}(ctxWithCancel, c, id, requiredIn)

	// Backup: NowNodes
	if c.NowNodes() != nil && c.Network() == MainNet {
		wg.Add(1)
		go func(ctx context.Context, client *Client, id string, requiredIn RequiredIn) {
			defer wg.Done()
			if resp, err := queryNowNodes(
				ctx, client, client.NowNodes(), id,
			); err == nil && checkRequirement(requiredIn, id, resp) {
				resultsChannel <- resp
			}
		}(ctxWithCancel, c, id, requiredIn)
	}

	// Waiting for all requests to finish
	go func() {
		wg.Wait()
		close(resultsChannel)
	}()

	return <-resultsChannel
}

// queryMAPI will submit a query transaction request to a miner using mAPI
func queryMAPI(ctx context.Context, client ClientInterface, minerCraft minercraft.TransactionService,
	miner *minercraft.Miner, id string) (*TransactionInfo, error) {
	client.DebugLog("executing request in mapi using miner: " + miner.Name)
	if resp, err := minerCraft.QueryTransaction(ctx, miner, id); err != nil {
		client.DebugLog("error executing request in mapi using miner: " + miner.Name + " failed: " + err.Error())
		return nil, err
	} else if resp != nil && resp.Query.ReturnResult == mAPISuccess && strings.EqualFold(resp.Query.TxID, id) {
		return &TransactionInfo{
			BlockHash:     resp.Query.BlockHash,
			BlockHeight:   resp.Query.BlockHeight,
			Confirmations: resp.Query.Confirmations,
			ID:            resp.Query.TxID,
			MinerID:       resp.Query.MinerID,
			Provider:      miner.Name,
		}, nil
	}
	return nil, ErrTransactionIDMismatch
}

// queryWhatsOnChain will request WhatsOnChain for transaction information
func queryWhatsOnChain(ctx context.Context, client ClientInterface,
	whatsOnChain whatsonchain.TransactionService, id string) (*TransactionInfo, error) {
	client.DebugLog("executing request in whatsonchain")
	if resp, err := whatsOnChain.GetTxByHash(ctx, id); err != nil {
		client.DebugLog("error executing request in whatsonchain: " + err.Error())
		return nil, err
	} else if resp != nil && strings.EqualFold(resp.TxID, id) {
		return &TransactionInfo{
			BlockHash:     resp.BlockHash,
			BlockHeight:   resp.BlockHeight,
			Confirmations: resp.Confirmations,
			ID:            resp.TxID,
			Provider:      providerWhatsOnChain,
			MinerID:       "",
		}, nil
	}
	return nil, ErrTransactionIDMismatch
}

// queryMatterCloud will request MatterCloud for transaction information
func queryMatterCloud(ctx context.Context, client ClientInterface,
	matterCloud mattercloud.TransactionService, id string) (*TransactionInfo, error) {
	client.DebugLog("executing request in mattercloud")
	if resp, err := matterCloud.Transaction(ctx, id); err != nil {
		client.DebugLog("error executing request in mattercloud: " + err.Error())
		return nil, err
	} else if resp != nil && strings.EqualFold(resp.TxID, id) {
		return &TransactionInfo{
			BlockHash:     resp.BlockHash,
			BlockHeight:   resp.BlockHeight,
			Confirmations: resp.Confirmations,
			ID:            resp.TxID,
			Provider:      providerMatterCloud,
			MinerID:       "",
		}, nil
	}
	return nil, ErrTransactionIDMismatch
}

// queryNowNodes will request NowNodes for transaction information
func queryNowNodes(ctx context.Context, client ClientInterface,
	nowNodes nownodes.TransactionService, id string) (*TransactionInfo, error) {
	client.DebugLog("executing request in nownodes")
	if resp, err := nowNodes.GetTransaction(ctx, nownodes.BSV, id); err != nil {
		client.DebugLog("error executing request in nownodes: " + err.Error())
		return nil, err
	} else if resp != nil && strings.EqualFold(resp.TxID, id) {
		return &TransactionInfo{
			BlockHash:     resp.BlockHash,
			BlockHeight:   resp.BlockHeight,
			Confirmations: resp.Confirmations,
			ID:            resp.TxID,
			Provider:      providerNowNodes,
			MinerID:       "",
		}, nil
	}
	return nil, ErrTransactionIDMismatch
}
