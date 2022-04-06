package bux

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
)

// PaymailDefaultServiceProvider is an interface for overriding the paymail actions in go-paymail/server
//
// This is an example and the default functionality for all the basic Paymail actions
type PaymailDefaultServiceProvider struct {
	client ClientInterface // (pointer) to the Client for accessing BUX model methods & etc
}

// createMetadata will create a new metadata seeded from the server information
func (p *PaymailDefaultServiceProvider) createMetadata(serverMetaData *server.RequestMetadata, request string) (metadata Metadata) {
	metadata = make(Metadata)
	metadata["paymail_request"] = request

	if serverMetaData != nil {
		if serverMetaData.UserAgent != "" {
			metadata["user_agent"] = serverMetaData.UserAgent
		}
		if serverMetaData.Note != "" {
			metadata["note"] = serverMetaData.Note
		}
		if serverMetaData.Domain != "" {
			metadata[domainField] = serverMetaData.Domain
		}
		if serverMetaData.IPAddress != "" {
			metadata["ip_address"] = serverMetaData.IPAddress
		}
	}
	return
}

// GetPaymailByAlias will get a paymail address and information by alias
func (p *PaymailDefaultServiceProvider) GetPaymailByAlias(ctx context.Context, alias, domain string,
	requestMetadata *server.RequestMetadata) (*paymail.AddressInformation, error) {

	// Create the metadata
	metadata := p.createMetadata(requestMetadata, "GetPaymailByAlias")

	// Create the paymail information
	paymailAddress, pubKey, destination, err := p.createPaymailInformation(
		ctx, alias, domain, append(p.client.DefaultModelOptions(), WithMetadatas(metadata))...,
	)
	if err != nil {
		return nil, err
	}

	// Return the information required by go-paymail
	return &paymail.AddressInformation{
		Alias:       paymailAddress.Alias,
		Avatar:      paymailAddress.Avatar,
		Domain:      paymailAddress.Domain,
		ID:          paymailAddress.ID,
		LastAddress: destination.Address,
		Name:        paymailAddress.Username,
		PubKey:      pubKey,
	}, nil
}

// CreateAddressResolutionResponse will create the address resolution response
func (p *PaymailDefaultServiceProvider) CreateAddressResolutionResponse(ctx context.Context, alias, domain string,
	_ bool, requestMetadata *server.RequestMetadata) (*paymail.ResolutionPayload, error) {

	// Create the metadata
	metadata := p.createMetadata(requestMetadata, "CreateAddressResolutionResponse")

	// Create the paymail information
	_, _, destination, err := p.createPaymailInformation(
		ctx, alias, domain, append(p.client.DefaultModelOptions(), WithMetadatas(metadata))...,
	)
	if err != nil {
		return nil, err
	}

	// Create the address resolution payload response
	return &paymail.ResolutionPayload{
		Address:   destination.Address,
		Output:    destination.LockingScript,
		Signature: "", // todo: add the signature if senderValidation is enabled
	}, nil
}

// CreateP2PDestinationResponse will create a p2p destination response
func (p *PaymailDefaultServiceProvider) CreateP2PDestinationResponse(ctx context.Context, alias, domain string,
	satoshis uint64, requestMetadata *server.RequestMetadata) (*paymail.PaymentDestinationPayload, error) {

	// Generate a unique reference ID
	referenceID, err := utils.RandomHex(16)
	if err != nil {
		return nil, err
	}

	// Create the metadata
	metadata := p.createMetadata(requestMetadata, "CreateP2PDestinationResponse")
	metadata[ReferenceIDField] = referenceID
	metadata[satoshisField] = satoshis

	// Create the paymail information
	// todo: strategy to break apart outputs based on satoshis (return x Outputs)
	var destination *Destination
	_, _, destination, err = p.createPaymailInformation(
		ctx, alias, domain, append(p.client.DefaultModelOptions(), WithMetadatas(metadata))...,
	)
	if err != nil {
		return nil, err
	}

	// Append the output(s)
	var outputs []*paymail.PaymentOutput
	outputs = append(outputs, &paymail.PaymentOutput{
		Address:  destination.Address,
		Satoshis: satoshis,
		Script:   destination.LockingScript,
	})

	return &paymail.PaymentDestinationPayload{
		Outputs:   outputs,
		Reference: referenceID,
	}, nil
}

// RecordTransaction will record the transaction
func (p *PaymailDefaultServiceProvider) RecordTransaction(ctx context.Context,
	p2pTx *paymail.P2PTransaction, requestMetadata *server.RequestMetadata) (*paymail.P2PTransactionPayload, error) {

	// Create the metadata
	metadata := p.createMetadata(requestMetadata, "RecordTransaction")
	metadata[p2pMetadataField] = p2pTx.MetaData
	metadata[ReferenceIDField] = p2pTx.Reference

	// todo: check if tx already exists, then gracefully respond?

	// Record the transaction
	transaction, err := p.client.RecordTransaction(
		ctx, "", p2pTx.Hex, "", []ModelOps{WithMetadatas(metadata)}...,
	)
	if err != nil {
		return nil, err
	}

	// Return the response from the p2p request
	return &paymail.P2PTransactionPayload{
		Note: p2pTx.MetaData.Note,
		TxID: transaction.ID,
	}, nil
}

// createPaymailInformation will get & create the paymail information (dynamic addresses)
func (p *PaymailDefaultServiceProvider) createPaymailInformation(ctx context.Context, alias, domain string,
	opts ...ModelOps) (paymailAddress *PaymailAddress, pubKey string, destination *Destination, err error) {

	// Get the paymail address record
	paymailAddress, err = getPaymail(ctx, alias+"@"+domain, opts...)
	if err != nil {
		return nil, "", nil, err
	}

	// Create the lock and set the release for after the function completes
	var unlock func()
	unlock, err = newWaitWriteLock(
		ctx, fmt.Sprintf(lockKeyProcessXpub, paymailAddress.XpubID), p.client.Cachestore(),
	)
	defer unlock()
	if err != nil {
		return nil, "", nil, err
	}

	// Get the corresponding xPub related to the paymail address
	var xPub *Xpub
	if xPub, err = getXpubByID(
		ctx, paymailAddress.XpubID, opts...,
	); err != nil {
		return nil, "", nil, err
	}

	// Get the external key (decrypted if needed)
	var externalXpub *bip32.ExtendedKey
	if externalXpub, err = paymailAddress.GetExternalXpub(); err != nil {
		return nil, "", nil, err
	}

	// Generate the new xPub and address with locking script
	var lockingScript string
	pubKey, _, lockingScript, err = getPaymailKeyInfo(
		externalXpub.String(),
		xPub.NextExternalNum,
	)
	if err != nil {
		return nil, "", nil, err
	}

	// create a new destination, based on the External xPub child
	// this is not yet possible within this library, it needs the full xPub
	destination = newDestination(paymailAddress.XpubID, lockingScript, append(opts, New())...)
	destination.Chain = utils.ChainExternal
	destination.Num = xPub.NextExternalNum

	// Create the new destination
	if err = destination.Save(ctx); err != nil {
		return nil, "", nil, err
	}

	// Increment and save
	xPub.NextExternalNum++
	if err = xPub.Save(ctx); err != nil {
		return nil, "", nil, err
	}
	return
}

// getPaymailKeyInfo will get all the paymail key information
func getPaymailKeyInfo(rawXPubKey string, num uint32) (pubKey, address, lockingScript string, err error) {

	// Get the xPub from string
	var hdKey *bip32.ExtendedKey
	hdKey, err = utils.ValidateXPub(rawXPubKey)
	if err != nil {
		return
	}

	// Get the child key
	var derivedKey *bip32.ExtendedKey
	if derivedKey, err = bitcoin.GetHDKeyChild(hdKey, num); err != nil {
		return
	}

	// Get the next key
	var nextKey *bec.PublicKey
	if nextKey, err = derivedKey.ECPubKey(); err != nil {
		return
	}
	pubKey = hex.EncodeToString(nextKey.SerialiseCompressed())

	// Get the address from the xPub
	var bsvAddress *bscript.Address
	if bsvAddress, err = bitcoin.GetAddressFromPubKey(
		nextKey, true,
	); err != nil {
		return
	}
	address = bsvAddress.AddressString

	// Generate a locking script for the address
	lockingScript, err = bitcoin.ScriptFromAddress(address)
	return
}
