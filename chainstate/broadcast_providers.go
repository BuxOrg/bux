package chainstate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mrz1836/go-nownodes"
	"github.com/tonicpow/go-minercraft/v2"
)

// generic broadcast provider
type txBroadcastProvider interface {
	getName() string
	broadcast(ctx context.Context, c *Client) error
}

// mAPI provider
type mapiBroadcastProvider struct {
	miner       *Miner
	txID, txHex string
}

func (provider mapiBroadcastProvider) getName() string {
	return provider.miner.Miner.Name
}

func (provider mapiBroadcastProvider) broadcast(ctx context.Context, c *Client) error {
	return broadcastMAPI(ctx, c, provider.miner.Miner, provider.txID, provider.txHex)
}

// broadcastMAPI will broadcast a transaction to a miner using mAPI
func broadcastMAPI(ctx context.Context, client ClientInterface, miner *minercraft.Miner, id, hex string) error {
	debugLog(client, id, "executing broadcast request in mapi using miner: "+miner.Name)

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
		debugLog(client, id, "error executing request in mapi using miner: "+miner.Name+" failed: "+err.Error())
		return err
	}

	// Something went wrong - got back an id that does not match
	if resp == nil || !strings.EqualFold(resp.Results.TxID, id) {
		return incorrectTxIDReturnedErr(resp.Results.TxID, id)
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
	txID, txHex string
}

func (provider whatsOnChainBroadcastProvider) getName() string {
	return ProviderWhatsOnChain
}

func (provider whatsOnChainBroadcastProvider) broadcast(ctx context.Context, c *Client) error {
	return broadcastWhatsOnChain(ctx, c, provider.txID, provider.txHex)
}

// broadcastWhatsOnChain will broadcast a transaction to WhatsOnChain
func broadcastWhatsOnChain(ctx context.Context, client ClientInterface, id, hex string) error {
	debugLog(client, id, "executing broadcast request for "+ProviderWhatsOnChain)

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
		return incorrectTxIDReturnedErr(txID, id)
	}

	// Success
	return nil
}

////

// NowNodes provider
type nowNodesBroadcastProvider struct {
	uniqueID, txID, txHex string
}

func (provider nowNodesBroadcastProvider) getName() string {
	return ProviderNowNodes
}

// Broadcast using NowNodes
func (provider nowNodesBroadcastProvider) broadcast(ctx context.Context, c *Client) error {
	return broadcastNowNodes(ctx, c, provider.uniqueID, provider.txID, provider.txHex)
}

// broadcastNowNodes will broadcast a transaction to NowNodes
func broadcastNowNodes(ctx context.Context, client ClientInterface, uniqueID, txID, hex string) error {
	debugLog(client, txID, "executing broadcast request for "+ProviderNowNodes)

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
		return incorrectTxIDReturnedErr(result.Result, txID)
	}

	// Success
	return nil
}

////

func incorrectTxIDReturnedErr(actualTxID, expectedTxID string) error {
	return fmt.Errorf("returned tx id [%s] does not match given tx id [%s]", actualTxID, expectedTxID)
}
