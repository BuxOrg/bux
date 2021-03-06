package chainstate

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-nownodes"
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

	// Set the error messages
	var errorMessages []string

	// First: try all mAPI miners (Only supported on main and test right now)
	if !utils.StringInSlice(ProviderMAPI, c.options.config.excludedProviders) {
		if c.Network() == MainNet || c.Network() == TestNet {
			for index := range c.options.config.mAPI.broadcastMiners {
				if c.options.config.mAPI.broadcastMiners[index] != nil {

					// Broadcast using mAPI
					provider = c.options.config.mAPI.broadcastMiners[index].Miner.Name
					if err = broadcastMAPI(
						ctxWithCancel, c, c.options.config.mAPI.broadcastMiners[index].Miner, id, hex,
					); err == nil { // Success response!
						return
					}

					// Check error response for "questionable errors"
					if doesErrorContain(err.Error(), broadcastQuestionableErrors) {
						err = checkInMempool(ctx, c, id, err.Error(), timeout)
						return // Success, found in mempool (or on-chain)
					}

					// Provider error?
					errorMessages = storeErrorMessage(c, errorMessages, err.Error(), provider)
				}
			}
		}
	}

	// Try next provider: WhatsOnChain
	if !utils.StringInSlice(ProviderWhatsOnChain, c.options.config.excludedProviders) {
		provider = ProviderWhatsOnChain
		if err = broadcastWhatsOnChain(ctx, c, id, hex); err != nil {

			// Check error response for (TX FAILURE)
			if doesErrorContain(err.Error(), broadcastQuestionableErrors) {
				err = checkInMempool(ctx, c, id, err.Error(), timeout)
				return // Success, found in mempool (or on-chain)
			}

			// Provider error?
			errorMessages = storeErrorMessage(c, errorMessages, err.Error(), provider)
		} else { // Success!
			return
		}
	}

	// Try next provider: NowNodes
	if !utils.StringInSlice(ProviderNowNodes, c.options.config.excludedProviders) {
		if c.NowNodes() != nil { // Only if NowNodes is loaded (requires API key)
			provider = ProviderNowNodes
			if err = broadcastNowNodes(ctx, c, id, id, hex); err != nil {

				// Check error response for (TX FAILURE)
				if doesErrorContain(err.Error(), broadcastQuestionableErrors) {
					err = checkInMempool(ctx, c, id, err.Error(), timeout)
					return // Success, found in mempool (or on-chain)
				}

				// Provider error?
				errorMessages = storeErrorMessage(c, errorMessages, err.Error(), provider)
			} else { // Success!
				return
			}
		}
	}

	// Final error - all failures
	return ProviderAll, errors.New("broadcast failed on all providers, errors: " + strings.Join(errorMessages, ","))
}

// storeErrorMessage will append the error and log it out
func storeErrorMessage(client ClientInterface, errorMessages []string, errorMessage, provider string) []string {
	errorMessages = append(errorMessages, provider+": "+errorMessage)
	client.DebugLog("broadcast error: " + errorMessage + " from provider: " + provider)
	return errorMessages
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
func broadcastMAPI(ctx context.Context, client ClientInterface, miner *minercraft.Miner, id, hex string) error {

	client.DebugLog("executing broadcast request in mapi using miner: " + miner.Name)

	resp, err := client.Minercraft().SubmitTransaction(ctx, miner, &minercraft.Transaction{
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

// broadcastWhatsOnChain will broadcast a transaction to WhatsOnChain
func broadcastWhatsOnChain(ctx context.Context, client ClientInterface, id, hex string) error {
	client.DebugLog("executing broadcast request for " + ProviderWhatsOnChain)

	txID, err := client.WhatsOnChain().BroadcastTx(ctx, hex)
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
func broadcastNowNodes(ctx context.Context, client ClientInterface, uniqueID, txID, hex string) error {
	client.DebugLog("executing broadcast request for " + ProviderNowNodes)

	result, err := client.NowNodes().SendRawTransaction(ctx, nownodes.BSV, hex, uniqueID)
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
