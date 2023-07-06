package chainstate

import (
	"context"
	"errors"
	"strings"
	"sync"
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
// NOTE: function register fastest successful broadcast into 'completeChannel' so client doesn't need to wait for other providers
func (c *Client) broadcast(ctx context.Context, id, hex string, timeout time.Duration, completeChannel, errorChannel chan string) {
	// Create a context (to cancel or timeout)
	ctxWithCancel, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wg sync.WaitGroup

	resultsChannel := make(chan broadcastResult)
	status := broadcastStatus{complete: false, syncChannel: completeChannel}

	// First: try all mAPI miners (Only supported on main and test right now)
	if shouldBroadcastWithMAPI(c) {
		for index := range c.options.config.mAPI.broadcastMiners { // why not for _, miner := range (...) ?
			miner := c.options.config.mAPI.broadcastMiners[index]
			if miner != nil {
				wg.Add(1)
				go func(miner *Miner) {
					defer wg.Done()

					provider := mapiBroadcastProvider{
						miner: miner,
						id:    id,
						hex:   hex,
					}

					broadcastToProvider(provider, id,
						c, ctxWithCancel, ctx, timeout,
						resultsChannel, &status)
				}(miner)
			}
		}
	}

	// Try next provider: WhatsOnChain
	if shouldBroadcastToWhatsOnChain(c) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			provider := whatsOnChainBroadcastProvider{
				id:  id,
				hex: hex,
			}

			broadcastToProvider(provider, id,
				c, ctx /* why ctx without timeout? */, ctx, timeout,
				resultsChannel, &status)
		}()
	}

	// Try next provider: NowNodes
	if shouldBroadcastToNowNodes(c) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			provider := nowNodesBroadcastProvider{
				uniqueID: id,
				txID:     id,
				hex:      hex,
			}

			broadcastToProvider(provider, id,
				c, ctx /* why ctx without timeout? */, ctx, timeout,
				resultsChannel, &status)
		}()
	}

	go func() {
		wg.Wait()
		close(resultsChannel)
		status.dispose()
	}()

	var errorMessages []string
	for result := range resultsChannel {
		if result.isError {
			errorMessages = storeErrorMessage(c, errorMessages, result.err.Error(), result.provider)
		}
		// log smth on success?
		// successProviders = append(successProviders, result.provider)
	}

	if !status.success && len(errorMessages) > 0 {
		errorChannel <- strings.Join(errorMessages, ", ")
	}
}

func shouldBroadcastWithMAPI(c *Client) bool {
	return !utils.StringInSlice(ProviderMAPI, c.options.config.excludedProviders) &&
		(c.Network() == MainNet || c.Network() == TestNet)
}

func shouldBroadcastToWhatsOnChain(c *Client) bool {
	return !utils.StringInSlice(ProviderWhatsOnChain, c.options.config.excludedProviders)
}

func shouldBroadcastToNowNodes(c *Client) bool {
	return !utils.StringInSlice(ProviderNowNodes, c.options.config.excludedProviders) &&
		c.NowNodes() != nil // Only if NowNodes is loaded (requires API key)
}

// storeErrorMessage will append the error and log it out
func storeErrorMessage(client ClientInterface, errorMessages []string, errorMessage, provider string) []string {
	errorMessages = append(errorMessages, provider+": "+errorMessage)
	client.DebugLog("broadcast error: " + errorMessage + " from provider: " + provider)
	return errorMessages
}

// struct handle communication with client - returns first successful broadcast
type broadcastStatus struct {
	mu          *sync.Mutex
	complete    bool
	success     bool
	syncChannel chan string
}

func (g *broadcastStatus) tryCompleteWithSuccess(fastestProvider string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.complete {
		g.complete = true
		g.success = true

		g.syncChannel <- fastestProvider
		close(g.syncChannel)
	}

	// g.mu.Unlock() is done by defer
}

func (g *broadcastStatus) dispose() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.complete {
		g.complete = true
		close(g.syncChannel) // have to close the channel here to not block client
	}

	// g.mu.Unlock() is done by defer
}

// result of single broadcast to provider
type broadcastResult struct {
	isError  bool
	err      error
	provider string
}

func newErrorResult(err error, provider string) broadcastResult {
	return broadcastResult{isError: true, err: err, provider: provider}
}

func newSuccessResult(provider string) broadcastResult {
	return broadcastResult{isError: false, provider: provider}
}

// generic broadcast provider
type txBroadcastProvider interface {
	getName() string
	broadcast(ctx context.Context, c *Client) error
}

func broadcastToProvider(provider txBroadcastProvider, id string,
	c *Client, ctx, fallbackCtx context.Context, timeout time.Duration,
	resultsChannel chan broadcastResult, status *broadcastStatus,
) {
	bErr := provider.broadcast(ctx, c)

	if bErr != nil {
		// check in Mempool as fallback - if transaction is there -> GREAT SUCCESS
		// Check error response for "questionable errors"/(TX FAILURE)
		if doesErrorContain(bErr.Error(), broadcastQuestionableErrors) {
			bErr = checkInMempool(fallbackCtx, c, id, bErr.Error(), timeout)
		}

		if bErr != nil {
			resultsChannel <- newErrorResult(bErr, provider.getName())
		}
	}

	// successful broadcast or found in mempool
	if bErr == nil {
		status.tryCompleteWithSuccess(provider.getName())
		resultsChannel <- newSuccessResult(provider.getName())
	}
}

type mapiBroadcastProvider struct {
	miner   *Miner
	id, hex string
}

func (provider mapiBroadcastProvider) getName() string {
	return provider.miner.Miner.Name
}

// Broadcast using mAPI
func (provider mapiBroadcastProvider) broadcast(ctx context.Context, c *Client) error {
	return broadcastMAPI(ctx, c, provider.miner.Miner, provider.id, provider.hex)
}

type whatsOnChainBroadcastProvider struct {
	id, hex string
}

func (provider whatsOnChainBroadcastProvider) getName() string {
	return ProviderWhatsOnChain
}

// Broadcast using WhatsOnChain
func (provider whatsOnChainBroadcastProvider) broadcast(ctx context.Context, c *Client) error {
	return broadcastWhatsOnChain(ctx, c, provider.id, provider.hex)
}

type nowNodesBroadcastProvider struct {
	uniqueID, txID, hex string
}

func (provider nowNodesBroadcastProvider) getName() string {
	return ProviderNowNodes
}

// Broadcast using NowNodes
func (provider nowNodesBroadcastProvider) broadcast(ctx context.Context, c *Client) error {
	return broadcastNowNodes(ctx, c, provider.uniqueID, provider.txID, provider.hex)
}

// checkInMempool is a quick check to see if the tx is in mempool (or on-chain)
func checkInMempool(ctx context.Context, client ClientInterface, id, errorMessage string, timeout time.Duration) error {
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
