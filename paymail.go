package bux

import (
	"context"
	"errors"
	"time"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/tonicpow/go-paymail"
)

// getCapabilities is a utility function to retrieve capabilities for a Paymail provider
func getCapabilities(ctx context.Context, cache cachestore.ClientInterface, client paymail.ClientInterface,
	domain string) (*paymail.CapabilitiesPayload, error) {

	// Attempt to get from cachestore
	// todo: allow user to configure the time that they want to cache the capabilities (if they want to cache or not)
	capabilities := new(paymail.CapabilitiesPayload)
	if err := cache.GetModel(
		ctx, cacheKeyCapabilities+domain, capabilities,
	); err != nil && !errors.Is(err, cachestore.ErrKeyNotFound) {
		return nil, err
	} else if capabilities != nil && len(capabilities.Capabilities) > 0 {
		return capabilities, nil
	}

	// Get SRV record (domain can be different!)
	srv, err := client.GetSRVRecord(
		paymail.DefaultServiceName, paymail.DefaultProtocol, domain,
	)
	if err != nil {
		return nil, err
	}

	// Get the capabilities
	var response *paymail.CapabilitiesResponse
	if response, err = client.GetCapabilities(
		srv.Target, int(srv.Port),
	); err != nil {
		return nil, err
	}

	// Save to cachestore
	if cache != nil {
		go func(cache cachestore.ClientInterface, key string, model *paymail.CapabilitiesPayload) {
			_ = cache.SetModel(ctx, key, model, cacheTTLCapabilities)
		}(cache, cacheKeyCapabilities+domain, &response.CapabilitiesPayload)
	}

	return &response.CapabilitiesPayload, nil
}

// hasP2P will return the P2P urls and true if they are both found
func hasP2P(capabilities *paymail.CapabilitiesPayload) (success bool, p2pDestinationURL, p2pSubmitTxURL string) {
	p2pDestinationURL = capabilities.GetString(paymail.BRFCP2PPaymentDestination, "")
	p2pSubmitTxURL = capabilities.GetString(paymail.BRFCP2PTransactions, "")

	if len(p2pSubmitTxURL) > 0 && len(p2pDestinationURL) > 0 {
		success = true
	}
	return
}

// resolvePaymailAddress is an old way to resolve a Paymail address (if P2P is not supported)
//
// Deprecated: this is already deprecated by TSC, use P2P or the new P4
func resolvePaymailAddress(ctx context.Context, cache cachestore.ClientInterface, client paymail.ClientInterface,
	capabilities *paymail.CapabilitiesPayload, alias, domain, purpose, senderPaymail string) (*paymail.ResolutionPayload, error) {

	// Attempt to get from cachestore
	// todo: allow user to configure the time that they want to cache the address resolution (if they want to cache or not)
	resolution := new(paymail.ResolutionPayload)
	if err := cache.GetModel(
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
	if cache != nil {
		go func(cache cachestore.ClientInterface, key string, model *paymail.ResolutionPayload) {
			_ = cache.SetModel(ctx, key, model, cacheTTLAddressResolution)
		}(cache, cacheKeyAddressResolution+alias+"-"+domain, &response.ResolutionPayload)
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
