package chainstate

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-nownodes"
	"github.com/tonicpow/go-minercraft"
)

// query will try ALL providers in order and return the first "valid" response based on requirements
func (c *Client) query(ctx context.Context, id string, requiredIn RequiredIn,
	timeout time.Duration) *TransactionInfo {

	// Create a context (to cancel or timeout)
	ctxWithCancel, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// First: try all mAPI miners (Only supported on main and test right now)
	if !utils.StringInSlice(ProviderMAPI, c.options.config.excludedProviders) {
		if c.Network() == MainNet || c.Network() == TestNet {
			for index := range c.options.config.mAPI.queryMiners {
				if c.options.config.mAPI.queryMiners[index] != nil {
					if res, err := queryMAPI(
						ctxWithCancel, c, c.options.config.mAPI.queryMiners[index].Miner, id,
					); err == nil && checkRequirement(requiredIn, id, res) {
						return res
					}
				}
			}
		}
	}

	// Next: try WhatsOnChain
	if !utils.StringInSlice(ProviderWhatsOnChain, c.options.config.excludedProviders) {
		if resp, err := queryWhatsOnChain(
			ctxWithCancel, c, id,
		); err == nil && checkRequirement(requiredIn, id, resp) {
			return resp
		}
	}

	// Next: try MatterCloud
	if !utils.StringInSlice(ProviderMatterCloud, c.options.config.excludedProviders) {
		if resp, err := queryMatterCloud(
			ctxWithCancel, c, id,
		); err == nil && checkRequirement(requiredIn, id, resp) {
			return resp
		}
	}

	// Next: try NowNodes (if loaded)
	if !utils.StringInSlice(ProviderNowNodes, c.options.config.excludedProviders) {
		if c.NowNodes() != nil && c.Network() == MainNet {
			if resp, err := queryNowNodes(
				ctxWithCancel, c, id,
			); err == nil && checkRequirement(requiredIn, id, resp) {
				return resp
			}
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
		// len(c.options.config.mAPI.queryMiners)+2,
	) // All miners & WhatsOnChain & MatterCloud

	// Create a context (to cancel or timeout)
	ctxWithCancel, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Loop each miner (break into a Go routine for each query)
	var wg sync.WaitGroup
	if !utils.StringInSlice(ProviderMAPI, c.options.config.excludedProviders) {
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
						ctx, client, miner, id,
					); err == nil && checkRequirement(requiredIn, id, res) {
						resultsChannel <- res
					}
				}(ctxWithCancel, c, &wg, c.options.config.mAPI.queryMiners[index].Miner, id, requiredIn)
			}
		}
	}

	// Backup: WhatsOnChain
	if !utils.StringInSlice(ProviderWhatsOnChain, c.options.config.excludedProviders) {
		wg.Add(1)
		go func(ctx context.Context, client *Client, id string, requiredIn RequiredIn) {
			defer wg.Done()
			if resp, err := queryWhatsOnChain(
				ctx, client, id,
			); err == nil && checkRequirement(requiredIn, id, resp) {
				resultsChannel <- resp
			}
		}(ctxWithCancel, c, id, requiredIn)
	}

	// Backup: MatterCloud
	if !utils.StringInSlice(ProviderMatterCloud, c.options.config.excludedProviders) {
		wg.Add(1)
		go func(ctx context.Context, client *Client, id string, requiredIn RequiredIn) {
			defer wg.Done()
			if resp, err := queryMatterCloud(
				ctx, client, id,
			); err == nil && checkRequirement(requiredIn, id, resp) {
				resultsChannel <- resp
			}
		}(ctxWithCancel, c, id, requiredIn)
	}

	// Backup: NowNodes
	if !utils.StringInSlice(ProviderNowNodes, c.options.config.excludedProviders) {
		if c.NowNodes() != nil && c.Network() == MainNet {
			wg.Add(1)
			go func(ctx context.Context, client *Client, id string, requiredIn RequiredIn) {
				defer wg.Done()
				if resp, err := queryNowNodes(
					ctx, client, id,
				); err == nil && checkRequirement(requiredIn, id, resp) {
					resultsChannel <- resp
				}
			}(ctxWithCancel, c, id, requiredIn)
		}
	}

	// Waiting for all requests to finish
	go func() {
		wg.Wait()
		close(resultsChannel)
	}()

	return <-resultsChannel
}

// queryMAPI will submit a query transaction request to a miner using mAPI
func queryMAPI(ctx context.Context, client ClientInterface, miner *minercraft.Miner, id string) (*TransactionInfo, error) {
	client.DebugLog("executing request in mapi using miner: " + miner.Name)
	if resp, err := client.Minercraft().QueryTransaction(ctx, miner, id); err != nil {
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
func queryWhatsOnChain(ctx context.Context, client ClientInterface, id string) (*TransactionInfo, error) {
	client.DebugLog("executing request in whatsonchain")
	if resp, err := client.WhatsOnChain().GetTxByHash(ctx, id); err != nil {
		client.DebugLog("error executing request in whatsonchain: " + err.Error())
		return nil, err
	} else if resp != nil && strings.EqualFold(resp.TxID, id) {
		return &TransactionInfo{
			BlockHash:     resp.BlockHash,
			BlockHeight:   resp.BlockHeight,
			Confirmations: resp.Confirmations,
			ID:            resp.TxID,
			Provider:      ProviderWhatsOnChain,
			MinerID:       "",
		}, nil
	}
	return nil, ErrTransactionIDMismatch
}

// queryMatterCloud will request MatterCloud for transaction information
func queryMatterCloud(ctx context.Context, client ClientInterface, id string) (*TransactionInfo, error) {
	client.DebugLog("executing request in mattercloud")
	if resp, err := client.MatterCloud().Transaction(ctx, id); err != nil {
		client.DebugLog("error executing request in mattercloud: " + err.Error())
		return nil, err
	} else if resp != nil && strings.EqualFold(resp.TxID, id) {
		return &TransactionInfo{
			BlockHash:     resp.BlockHash,
			BlockHeight:   resp.BlockHeight,
			Confirmations: resp.Confirmations,
			ID:            resp.TxID,
			Provider:      ProviderMatterCloud,
			MinerID:       "",
		}, nil
	}
	return nil, ErrTransactionIDMismatch
}

// queryNowNodes will request NowNodes for transaction information
func queryNowNodes(ctx context.Context, client ClientInterface, id string) (*TransactionInfo, error) {
	client.DebugLog("executing request in nownodes")
	if resp, err := client.NowNodes().GetTransaction(ctx, nownodes.BSV, id); err != nil {
		client.DebugLog("error executing request in nownodes: " + err.Error())
		return nil, err
	} else if resp != nil && strings.EqualFold(resp.TxID, id) {
		return &TransactionInfo{
			BlockHash:     resp.BlockHash,
			BlockHeight:   resp.BlockHeight,
			Confirmations: resp.Confirmations,
			ID:            resp.TxID,
			Provider:      ProviderNowNodes,
			MinerID:       "",
		}, nil
	}
	return nil, ErrTransactionIDMismatch
}
