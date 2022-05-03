package chainstate

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mrz1836/go-mattercloud"
	"github.com/mrz1836/go-nownodes"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/tonicpow/go-minercraft"
)

var (
	// broadcastSuccessErrors are a list of errors that are still considered a success
	broadcastSuccessErrors = []string{
		"already in the mempool", // {"error": "-27: Transaction already in the mempool"}
		"txn-already-know",       // { "error": "-26: 257: txn-already-known"}  // txn-already-know
		"txn-already-in-mempool", // txn-already-in-mempool
		"txn_already_known",      // TXN_ALREADY_KNOWN
		"txn_already_in_mempool", // TXN_ALREADY_IN_MEMPOOL
	}

	// broadcastQuestionableErrors are a list of errors that are not good broadcast responses,
	// but need to be checked differently
	broadcastQuestionableErrors = []string{
		"missing inputs", // Returned from mAPI for a valid tx that is on-chain
	}

	/*
		TXN_ALREADY_KNOWN (suppressed - returns as success: true)
		TXN_ALREADY_IN_MEMPOOL (suppressed - returns as success: true)
		TXN_MEMPOOL_CONFLICT
		NON_FINAL_POOL_FULL
		TOO_LONG_NON_FINAL_CHAIN
		BAD_TXNS_INPUTS_TOO_LARGE
		BAD_TXNS_INPUTS_SPENT
		NON_BIP68_FINAL
		TOO_LONG_VALIDATION_TIME
		BAD_TXNS_NONSTANDARD_INPUTS
		ABSURDLY_HIGH_FEE
		DUST
		TX_FEE_TOO_LOW
	*/
)

// doesErrorContain will look at a string for a list of strings
func doesErrorContain(err string, messages []string) bool {
	lower := strings.ToLower(err)
	for _, str := range messages {
		if strings.Contains(lower, str) {
			return true
		}
	}
	return false
}

// broadcast will broadcast using a standard strategy
//
// NOTE: if successful (in-mempool), no error will be returned
func (c *Client) broadcast(ctx context.Context, id, hex string, timeout time.Duration) (provider string, err error) {

	// Create a context (to cancel or timeout)
	ctxWithCancel, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// First: try all mAPI miners (Only supported on main and test right now)
	if c.Network() == MainNet || c.Network() == TestNet {
		for index := range c.options.config.mAPI.broadcastMiners {
			if c.options.config.mAPI.broadcastMiners[index] != nil {

				// Broadcast using mAPI
				provider = c.options.config.mAPI.broadcastMiners[index].Miner.Name
				err = broadcastMAPI(
					ctxWithCancel, c, c.Minercraft(),
					c.options.config.mAPI.broadcastMiners[index].Miner, id, hex,
				)
				if err == nil { // Success response!
					return
				}

				// Check error response for "questionable errors"
				if doesErrorContain(err.Error(), broadcastQuestionableErrors) {
					err = checkInMempool(ctx, c, id, err.Error(), timeout)
					return // Success, found in mempool (or on-chain)
				}

				// Provider error?
				c.DebugLog("broadcast error: " + err.Error() + " from provider: " + c.options.config.mAPI.broadcastMiners[index].Miner.Name)
			}
		}
	}

	// Try next provider: MatterCloud
	provider = providerMatterCloud
	if err = broadcastMatterCloud(ctx, c, c.MatterCloud(), id, hex); err != nil {

		// Check error response for (TX FAILURE)
		if doesErrorContain(err.Error(), broadcastQuestionableErrors) {
			err = checkInMempool(ctx, c, id, err.Error(), timeout)
			return // Success, found in mempool (or on-chain)
		}

		// Provider error?
		c.DebugLog("broadcast error: " + err.Error() + " from provider: " + providerMatterCloud)
	} else { // Success!
		return
	}

	// Try next provider: WhatsOnChain
	provider = providerWhatsOnChain
	if err = broadcastWhatsOnChain(ctx, c, c.WhatsOnChain(), id, hex); err != nil {

		// Check error response for (TX FAILURE)
		if doesErrorContain(err.Error(), broadcastQuestionableErrors) {
			err = checkInMempool(ctx, c, id, err.Error(), timeout)
			return // Success, found in mempool (or on-chain)
		}

		// Provider error?
		c.DebugLog("broadcast error: " + err.Error() + " from provider: " + providerWhatsOnChain)
	} else { // Success!
		return
	}

	// Try next provider: NowNodes
	provider = providerNowNodes
	if c.NowNodes() != nil { // Only if NowNodes is loaded (requires API key)
		if err = broadcastNowNodes(ctx, c, c.NowNodes(), id, id, hex); err != nil {

			// Check error response for (TX FAILURE)
			if doesErrorContain(err.Error(), broadcastQuestionableErrors) {
				err = checkInMempool(ctx, c, id, err.Error(), timeout)
				return // Success, found in mempool (or on-chain)
			}

			// Provider error?
			c.DebugLog("broadcast error: " + err.Error() + " from provider: " + providerNowNodes)
		} else { // Success!
			return
		}
	}

	// Final error - all failures
	return providerAll, errors.New("broadcast failed on all providers")
}

// checkInMempool is a quick check to see if the tx is in mempool (or on-chain)
func checkInMempool(ctx context.Context, client ClientInterface, id, errorMessage string,
	timeout time.Duration) error {
	if _, err := client.QueryTransaction(
		ctx, id, requiredInMempool, timeout,
	); err != nil {
		return errors.New("error query tx failed: " + err.Error() + ", " + "broadcast initial error: " + errorMessage)
	}
	return nil
}

// broadcastMAPI will broadcast a transaction to a miner using mAPI
func broadcastMAPI(ctx context.Context, client ClientInterface, minerCraft minercraft.TransactionService,
	miner *minercraft.Miner, id, hex string) error {

	client.DebugLog("executing broadcast request in mapi using miner: " + miner.Name)

	resp, err := minerCraft.SubmitTransaction(ctx, miner, &minercraft.Transaction{
		CallBackEncryption: "", // todo: allow customizing the payload
		CallBackToken:      "",
		CallBackURL:        "",
		DsCheck:            false,
		MerkleFormat:       "",
		MerkleProof:        false,
		RawTx:              hex,
	})
	if err != nil {
		client.DebugLog("error executing request in mapi using miner: " + miner.Name + " failed: " + err.Error())
		return err
	}

	// Something went wrong - got back an id that does not match
	if resp == nil || !strings.EqualFold(resp.Results.TxID, id) {
		return errors.New("returned tx id [" + resp.Results.TxID + "] does not match given tx id [" + id + "]")
	}

	// mAPI success of broadcast
	if resp.Results.ReturnResult == mAPISuccess {
		return nil
	}

	// Check error message (for success error message)
	if doesErrorContain(resp.Results.ResultDescription, broadcastSuccessErrors) {
		return nil
	}

	// We got a potential real error message?
	return errors.New(resp.Results.ResultDescription)
}

// broadcastMatterCloud will broadcast a transaction to MatterCloud
func broadcastMatterCloud(ctx context.Context, client ClientInterface, matterCloud mattercloud.TransactionService,
	id, hex string) error {
	client.DebugLog("executing broadcast request for " + providerMatterCloud)

	resp, err := matterCloud.Broadcast(ctx, hex)
	if err != nil {

		// Check error message (for success error message)
		if doesErrorContain(err.Error(), broadcastSuccessErrors) {
			return nil
		}
		return err
	}

	// Something went wrong - got back an id that does not match
	if !strings.EqualFold(resp.Result.TxID, id) {
		return errors.New("returned tx id [" + resp.Result.TxID + "] does not match given tx id [" + id + "]")
	}

	// Success was a failure?
	if !resp.Success {
		return errors.New("returned response was unsuccessful")
	}

	// Success
	return nil
}

// broadcastWhatsOnChain will broadcast a transaction to WhatsOnChain
func broadcastWhatsOnChain(ctx context.Context, client ClientInterface, whatsOnChain whatsonchain.TransactionService,
	id, hex string) error {
	client.DebugLog("executing broadcast request for " + providerWhatsOnChain)

	txID, err := whatsOnChain.BroadcastTx(ctx, hex)
	if err != nil {

		// Check error message (for success error message)
		if doesErrorContain(err.Error(), broadcastSuccessErrors) {
			return nil
		}
		return err
	}

	// Something went wrong - got back an id that does not match
	if !strings.EqualFold(txID, id) {
		return errors.New("returned tx id [" + txID + "] does not match given tx id [" + id + "]")
	}

	// Success
	return nil
}

// broadcastNowNodes will broadcast a transaction to NowNodes
func broadcastNowNodes(ctx context.Context, client ClientInterface, nowNodes nownodes.TransactionService,
	uniqueID, txID, hex string) error {
	client.DebugLog("executing broadcast request for " + providerNowNodes)

	result, err := nowNodes.SendRawTransaction(ctx, nownodes.BSV, hex, uniqueID)
	if err != nil {

		// Check error message (for success error message)
		if doesErrorContain(err.Error(), broadcastSuccessErrors) {
			return nil
		}
		return err
	}

	// Something went wrong - got back an id that does not match
	if !strings.EqualFold(result.Result, txID) {
		return errors.New("returned tx id [" + result.Result + "] does not match given tx id [" + txID + "]")
	}

	// Success
	return nil
}
