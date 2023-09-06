package bux

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mrz1836/go-cachestore"
	"github.com/tonicpow/go-paymail"
)

// getCapabilities is a utility function to retrieve capabilities for a Paymail provider
func getCapabilities(ctx context.Context, cs cachestore.ClientInterface, client paymail.ClientInterface,
	domain string) (*paymail.CapabilitiesPayload, error) {

	// Attempt to get from cachestore
	// todo: allow user to configure the time that they want to cache the capabilities (if they want to cache or not)
	capabilities := new(paymail.CapabilitiesPayload)
	if err := cs.GetModel(
		ctx, cacheKeyCapabilities+domain, capabilities,
	); err != nil && !errors.Is(err, cachestore.ErrKeyNotFound) {
		return nil, err
	} else if capabilities != nil && len(capabilities.Capabilities) > 0 {
		return capabilities, nil
	}

	// Get SRV record (domain can be different!)
	var response *paymail.CapabilitiesResponse
	srv, err := client.GetSRVRecord(
		paymail.DefaultServiceName, paymail.DefaultProtocol, domain,
	)
	if err != nil {
		// Error returned was a real error
		if !strings.Contains(err.Error(), "zero SRV records found") { // This error is from no SRV record being found
			return nil, err
		}

		// Try to get capabilities without the SRV record
		// 'Should no record be returned, a paymail client should assume a host of <domain>.<tld> and a port of 443.'
		// http://bsvalias.org/02-01-host-discovery.html

		// Get the capabilities via target
		if response, err = client.GetCapabilities(
			domain, paymail.DefaultPort,
		); err != nil {
			return nil, err
		}
	} else {
		// Get the capabilities via SRV record
		if response, err = client.GetCapabilities(
			srv.Target, int(srv.Port),
		); err != nil {
			return nil, err
		}
	}

	// Save to cachestore
	if cs != nil && !cs.Engine().IsEmpty() {
		_ = cs.SetModel(
			context.Background(), cacheKeyCapabilities+domain,
			&response.CapabilitiesPayload, cacheTTLCapabilities,
		)
	}

	return &response.CapabilitiesPayload, nil
}

// hasP2P will return the P2P urls and true if they are both found
func hasP2P(capabilities *paymail.CapabilitiesPayload) (success bool, p2pDestinationURL, p2pSubmitTxURL string, format PaymailPayloadFormat) {
	p2pDestinationURL = capabilities.GetString(paymail.BRFCP2PPaymentDestination, "")
	p2pSubmitTxURL = capabilities.GetString(paymail.BRFCP2PTransactions, "")
	p2pBeefSubmitTxURL := capabilities.GetString(paymail.BRFCBeefTransaction, "")

	if len(p2pBeefSubmitTxURL) > 0 {
		p2pSubmitTxURL = p2pBeefSubmitTxURL
		format = BeefPaymailPayloadFormat
	}
	//else {
	//	format = BasicPaymailPayloadFormat
	//}

	if len(p2pSubmitTxURL) > 0 && len(p2pDestinationURL) > 0 {
		success = true
	}
	return
}

// resolvePaymailAddress is an old way to resolve a Paymail address (if P2P is not supported)
//
// Deprecated: this is already deprecated by TSC, use P2P or the new P4
func resolvePaymailAddress(ctx context.Context, cs cachestore.ClientInterface, client paymail.ClientInterface,
	capabilities *paymail.CapabilitiesPayload, alias, domain, purpose, senderPaymail string) (*paymail.ResolutionPayload, error) {

	// Attempt to get from cachestore
	// todo: allow user to configure the time that they want to cache the address resolution (if they want to cache or not)
	resolution := new(paymail.ResolutionPayload)
	if err := cs.GetModel(
		ctx, cacheKeyAddressResolution+alias+"-"+domain, resolution,
	); err != nil && !errors.Is(err, cachestore.ErrKeyNotFound) {
		return nil, err
	} else if resolution != nil && len(resolution.Output) > 0 {
		return resolution, nil
	}

	// Get the URL
	addressResolutionURL := capabilities.GetString(
		paymail.BRFCBasicAddressResolution, paymail.BRFCPaymentDestination,
	)
	if len(addressResolutionURL) == 0 {
		return nil, ErrMissingAddressResolutionURL
	}

	// Resolve address
	response, err := client.ResolveAddress(
		addressResolutionURL,
		alias, domain,
		&paymail.SenderRequest{
			Dt:           time.Now().UTC().Format(time.RFC3339), // UTC is assumed
			Purpose:      purpose,                               // Generic message about the resolution
			SenderHandle: senderPaymail,                         // Assumed it's a paymail@domain.com
		},
	)
	if err != nil {
		return nil, err
	}

	// Save to cachestore
	if cs != nil && !cs.Engine().IsEmpty() {
		_ = cs.SetModel(
			ctx, cacheKeyAddressResolution+alias+"-"+domain,
			&response.ResolutionPayload, cacheTTLAddressResolution,
		)
	}

	return &response.ResolutionPayload, nil
}

// startP2PTransaction will start the P2P transaction, returning the reference ID and outputs
func startP2PTransaction(client paymail.ClientInterface,
	alias, domain, p2pDestinationURL string, satoshis uint64) (*paymail.PaymentDestinationPayload, error) {

	// Start the P2P transaction request
	response, err := client.GetP2PPaymentDestination(
		p2pDestinationURL,
		alias, domain,
		&paymail.PaymentRequest{Satoshis: satoshis},
	)
	if err != nil {
		return nil, err
	}

	return &response.PaymentDestinationPayload, nil
}

// finalizeP2PTransaction will notify the paymail provider about the transaction
func finalizeP2PTransaction(client paymail.ClientInterface,
	alias, domain, p2pSubmitURL, referenceID, note, senderPaymailAddress, txHex, txBeef string) (*paymail.P2PTransactionPayload, error) {

	// Submit the P2P transaction
	/*logger.Data(2, logger.DEBUG, "sending p2p tx...",
		logger.MakeParameter("alias", alias),
		logger.MakeParameter("p2pSubmitURL", p2pSubmitURL),
		logger.MakeParameter("domain", domain),
		logger.MakeParameter("note", note),
		logger.MakeParameter("senderPaymailAddress", senderPaymailAddress),
		logger.MakeParameter("referenceID", referenceID),
	)*/

	p2pTransaction := &paymail.P2PTransaction{
		MetaData: &paymail.P2PMetaData{
			Note:   note,
			Sender: senderPaymailAddress,
		},
		Reference: referenceID,
	}

	if len(txBeef) > 0 {
		p2pTransaction.Beef = txBeef
	} else {
		p2pTransaction.Hex = txHex
	}

	response, err := client.SendP2PTransaction(p2pSubmitURL, alias, domain, p2pTransaction)
	if err != nil {
		return nil, err
	}

	return &response.P2PTransactionPayload, nil
}
