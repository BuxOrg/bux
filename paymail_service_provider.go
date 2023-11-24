package bux

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/bitcoin-sv/go-paymail"
	"github.com/bitcoin-sv/go-paymail/server"
	"github.com/bitcoin-sv/go-paymail/spv"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	customTypes "github.com/mrz1836/go-datastore/custom_types"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/utils"
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
func (p *PaymailDefaultServiceProvider) GetPaymailByAlias(
	ctx context.Context,
	alias, domain string,
	requestMetadata *server.RequestMetadata,
) (*paymail.AddressInformation, error) {
	metadata := p.createMetadata(requestMetadata, "GetPaymailByAlias")

	paymailAddress, pubKey, err := p.createPaymailInformation(
		ctx, alias, domain, append(p.client.DefaultModelOptions(), WithMetadatas(metadata))...,
	)
	if err != nil {
		return nil, err
	}

	return &paymail.AddressInformation{
		Alias:  paymailAddress.Alias,
		Avatar: paymailAddress.Avatar,
		Domain: paymailAddress.Domain,
		ID:     paymailAddress.ID,
		Name:   paymailAddress.PublicName,
		PubKey: pubKey.pubKey,
	}, nil
}

// CreateAddressResolutionResponse will create the address resolution response
func (p *PaymailDefaultServiceProvider) CreateAddressResolutionResponse(
	ctx context.Context,
	alias, domain string,
	_ bool,
	requestMetadata *server.RequestMetadata,
) (*paymail.ResolutionPayload, error) {
	metadata := p.createMetadata(requestMetadata, "CreateAddressResolutionResponse")

	paymailAddress, pubKey, err := p.createPaymailInformation(
		ctx, alias, domain, append(p.client.DefaultModelOptions(), WithMetadatas(metadata))...,
	)
	if err != nil {
		return nil, err
	}
	destination, err := createDestination(
		ctx, paymailAddress, pubKey, true, append(p.client.DefaultModelOptions(), WithMetadatas(metadata))...,
	)
	if err != nil {
		return nil, err
	}

	return &paymail.ResolutionPayload{
		Address:   destination.Address,
		Output:    destination.LockingScript,
		Signature: "", // todo: add the signature if senderValidation is enabled
	}, nil
}

// CreateP2PDestinationResponse will create a p2p destination response
func (p *PaymailDefaultServiceProvider) CreateP2PDestinationResponse(
	ctx context.Context,
	alias, domain string,
	satoshis uint64,
	requestMetadata *server.RequestMetadata,
) (*paymail.PaymentDestinationPayload, error) {
	referenceID, err := utils.RandomHex(16)
	if err != nil {
		return nil, err
	}

	metadata := p.createMetadata(requestMetadata, "CreateP2PDestinationResponse")
	metadata[ReferenceIDField] = referenceID
	metadata[satoshisField] = satoshis

	// todo: strategy to break apart outputs based on satoshis (return x Outputs)
	var destination *Destination
	paymailAddress, pubKey, err := p.createPaymailInformation(
		ctx, alias, domain, append(p.client.DefaultModelOptions(), WithMetadatas(metadata))...,
	)
	if err != nil {
		return nil, err
	}
	destination, err = createDestination(
		ctx, paymailAddress, pubKey, false, append(p.client.DefaultModelOptions(), WithMetadatas(metadata))...,
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
// TODO: rename to HandleReceivedP2pTransaction
func (p *PaymailDefaultServiceProvider) RecordTransaction(ctx context.Context,
	p2pTx *paymail.P2PTransaction, requestMetadata *server.RequestMetadata) (*paymail.P2PTransactionPayload, error) {

	// Create the metadata
	metadata := p.createMetadata(requestMetadata, "HandleReceivedP2pTransaction")
	metadata[p2pMetadataField] = p2pTx.MetaData
	metadata[ReferenceIDField] = p2pTx.Reference

	// Record the transaction
	rts, err := getIncomingTxRecordStrategy(ctx, p.client, p2pTx.Hex)
	if err != nil {
		return nil, err
	}

	rts.ForceBroadcast(true)

	if p2pTx.Beef != "" {
		rts.FailOnBroadcastError(true)
	}

	transaction, err := recordTransaction(ctx, p.client, rts, WithMetadatas(metadata))
	if err != nil {
		return nil, err
	}

	if p2pTx.DecodedBeef == nil {
		return &paymail.P2PTransactionPayload{
			Note: p2pTx.MetaData.Note,
			TxID: transaction.ID,
		}, nil
	}

	if reflect.TypeOf(rts) == reflect.TypeOf(&externalIncomingTx{}) {
		for _, input := range p2pTx.DecodedBeef.InputsTxData {
			err = saveBeefTransactionInput(ctx, p.client, input.Transaction)
			if err != nil {
				p.client.Logger().Error(ctx, "error in saveBeefTransactionInput", err)
			}
		}
	}

	// Return the response from the p2p request
	return &paymail.P2PTransactionPayload{
		Note: p2pTx.MetaData.Note,
		TxID: transaction.ID,
	}, nil
}

// VerifyMerkleRoots will verify the merkle roots by checking them in external header service - Pulse
func (p *PaymailDefaultServiceProvider) VerifyMerkleRoots(
	ctx context.Context,
	merkleRoots []*spv.MerkleRootConfirmationRequestItem,
) error {
	request := make([]chainstate.MerkleRootConfirmationRequestItem, 0)
	for _, m := range merkleRoots {
		request = append(request, chainstate.MerkleRootConfirmationRequestItem{
			MerkleRoot:  m.MerkleRoot,
			BlockHeight: m.BlockHeight,
		})
	}
	return p.client.Chainstate().VerifyMerkleRoots(ctx, request)
}

func (p *PaymailDefaultServiceProvider) createPaymailInformation(ctx context.Context, alias, domain string, opts ...ModelOps) (paymailAddress *PaymailAddress, pubKey *derivedPubKey, err error) {
	paymailAddress, err = getPaymailAddress(ctx, alias+"@"+domain, opts...)
	if err != nil {
		return nil, nil, err
	} else if paymailAddress == nil {
		return nil, nil, ErrMissingPaymail
	}

	unlock, err := newWaitWriteLock(ctx, lockKey(paymailAddress), p.client.Cachestore())
	defer unlock()
	if err != nil {
		return nil, nil, err
	}

	xPub, err := getXpubForPaymail(ctx, p.client, paymailAddress, opts)
	if err != nil {
		return nil, nil, err
	}

	externalXpub, err := paymailAddress.GetExternalXpub()
	if err != nil {
		return nil, nil, err
	}

	chainNum, err := xPub.incrementNextNum(ctx, utils.ChainExternal)
	if err != nil {
		return nil, nil, err
	}

	pubKey, err = deriveKey(externalXpub.String(), chainNum)
	if err != nil {
		return nil, nil, err
	}
	return
}

func getXpubForPaymail(ctx context.Context, client ClientInterface, paymailAddress *PaymailAddress, opts []ModelOps) (*Xpub, error) {
	return getXpubWithCache(
		ctx, client, "", paymailAddress.XpubID, opts...,
	)
}

func createDestination(ctx context.Context, paymailAddress *PaymailAddress, pubKey *derivedPubKey, monitor bool, opts ...ModelOps) (destination *Destination, err error) {
	lockingScript, err := createLockingScript(pubKey.ecPubKey)
	if err != nil {
		return nil, err
	}

	// create a new destination, based on the External xPub child
	// this is not yet possible using the xpub struct. That needs the full xPub, which we don't have.
	destination = newDestination(paymailAddress.XpubID, lockingScript, append(opts, New())...)
	destination.Chain = utils.ChainExternal
	destination.Num = pubKey.chainNum

	// Only on for basic address resolution, not enabled for p2p
	if monitor {
		destination.Monitor = customTypes.NullTime{NullTime: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		}}
	}

	if err = destination.Save(ctx); err != nil {
		return nil, err
	}

	return
}

func lockKey(paymailAddress *PaymailAddress) string {
	return fmt.Sprintf(lockKeyProcessXpub, paymailAddress.XpubID)
}

func createLockingScript(ecPubKey *bec.PublicKey) (lockingScript string, err error) {
	bsvAddress, err := bitcoin.GetAddressFromPubKey(ecPubKey, true)
	if err != nil {
		return
	}
	address := bsvAddress.AddressString

	lockingScript, err = bitcoin.ScriptFromAddress(address)
	return
}

type derivedPubKey struct {
	ecPubKey *bec.PublicKey
	chainNum uint32
	pubKey   string
}

func deriveKey(rawXPubKey string, num uint32) (k *derivedPubKey, err error) {
	k = &derivedPubKey{chainNum: num}

	hdKey, err := utils.ValidateXPub(rawXPubKey)
	if err != nil {
		return
	}

	derivedKey, err := bitcoin.GetHDKeyChild(hdKey, num)
	if err != nil {
		return
	}

	k.ecPubKey, err = derivedKey.ECPubKey()
	if err != nil {
		return
	}

	k.pubKey = hex.EncodeToString(k.ecPubKey.SerialiseCompressed())
	return
}
