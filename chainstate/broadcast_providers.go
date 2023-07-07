package chainstate

import (
	"context"
	"errors"
	"strings"

	"github.com/mrz1836/go-nownodes"
	"github.com/tonicpow/go-minercraft"
)

// generic broadcast provider
type txBroadcastProvider interface {
	getName() string
	broadcast(ctx context.Context, c *Client) error
}

// mAPI provider
type mapiBroadcastProvider struct {
	miner   *Miner
	id, hex string
}

func (provider mapiBroadcastProvider) getName() string {
	return provider.miner.Miner.Name
}

func (provider mapiBroadcastProvider) broadcast(ctx context.Context, c *Client) error {
	return broadcastMAPI(ctx, c, provider.miner.Miner, provider.id, provider.hex)
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

////

// WhatsOnChain provider
type whatsOnChainBroadcastProvider struct {
	id, hex string
}

func (provider whatsOnChainBroadcastProvider) getName() string {
	return ProviderWhatsOnChain
}

func (provider whatsOnChainBroadcastProvider) broadcast(ctx context.Context, c *Client) error {
	return broadcastWhatsOnChain(ctx, c, provider.id, provider.hex)
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

////

// NowNodes provider
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

////
