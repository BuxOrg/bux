package bux

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
	xtester "github.com/BuxOrg/bux/tester"
	"github.com/jarcoal/httpmock"
	"github.com/mrz1836/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
)

const (
	testAlias     = "tester"
	testDomain    = "test.com"
	testServerURL = "https://" + testDomain + "/api/v1/" + paymail.DefaultServiceName
	testOutput    = "76a9147f11c8f67a2781df0400ebfb1f31b4c72a780b9d88ac"
)

// newTestPaymailClient will return a client for testing purposes
func newTestPaymailClient(t *testing.T, domains []string) paymail.ClientInterface {
	newClient, err := xtester.PaymailMockClient(domains)
	require.NotNil(t, newClient)
	require.NoError(t, err)
	return newClient
}

// newTestPaymailConfig loads a basic test configuration
func newTestPaymailConfig(t *testing.T, domain string) *server.Configuration {
	c, err := server.NewConfig(
		new(mockServiceProvider),
		server.WithDomain(domain),
		server.WithP2PCapabilities(),
	)
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

// mockValidResponse is used for mocking the response
func mockValidResponse(statusCode int, p2p bool, domain string) {
	httpmock.Reset()

	serverURL := "https://" + domain + "/api/v1/" + paymail.DefaultServiceName

	// Basic address resolution vs P2P
	if !p2p {
		httpmock.RegisterResponder(http.MethodGet, "https://"+domain+":443/.well-known/"+paymail.DefaultServiceName,
			httpmock.NewStringResponder(
				statusCode,
				`{"`+paymail.DefaultServiceName+`": "`+paymail.DefaultBsvAliasVersion+`","capabilities":{
"`+paymail.BRFCSenderValidation+`": false,
"`+paymail.BRFCPki+`": "`+serverURL+`/id/{alias}@{domain.tld}",
"`+paymail.BRFCPaymentDestination+`": "`+serverURL+`/address/{alias}@{domain.tld}"}
}`,
			),
		)
	} else {
		httpmock.RegisterResponder(http.MethodGet, "https://"+domain+":443/.well-known/"+paymail.DefaultServiceName,
			httpmock.NewStringResponder(
				statusCode,
				`{"`+paymail.DefaultServiceName+`": "`+paymail.DefaultBsvAliasVersion+`","capabilities":{
"`+paymail.BRFCSenderValidation+`": false,
"`+paymail.BRFCPki+`": "`+serverURL+`/id/{alias}@{domain.tld}",
"`+paymail.BRFCPaymentDestination+`": "`+serverURL+`/address/{alias}@{domain.tld}",
"`+paymail.BRFCP2PTransactions+`": "`+serverURL+`/receive-transaction/{alias}@{domain.tld}",
"`+paymail.BRFCP2PPaymentDestination+`": "`+serverURL+`/p2p-payment-destination/{alias}@{domain.tld}"}
}`,
			),
		)
	}

	httpmock.RegisterResponder(http.MethodPost, serverURL+"/p2p-payment-destination/"+testAlias+"@"+testDomain,
		httpmock.NewStringResponder(
			statusCode,
			`{"outputs": [{"script": "76a9143e2d1d795f8acaa7957045cc59376177eb04a3c588ac","satoshis": 1000}],"reference": "z0bac4ec-6f15-42de-9ef4-e60bfdabf4f7"}`,
		),
	)

	httpmock.RegisterResponder(http.MethodPost, serverURL+"/address/"+testAlias+"@"+domain,
		httpmock.NewStringResponder(
			statusCode,
			`{"output": "`+testOutput+`"}`,
		),
	)
}

// TestPaymailClient will test various Paymail client methods
func TestPaymailClient(t *testing.T) {
	t.Parallel()

	config := newTestPaymailConfig(t, testDomain)
	require.NotNil(t, config)

	client := newTestPaymailClient(t, []string{testDomain})
	require.NotNil(t, client)
}

// Test_hasP2P will test the method hasP2P()
func Test_hasP2P(t *testing.T) {
	t.Parallel()

	t.Run("no p2p capabilities", func(t *testing.T) {
		capabilities := server.GenericCapabilities(paymail.DefaultBsvAliasVersion, false)
		success, p2pDestinationURL, p2pSubmitTxURL := hasP2P(capabilities)
		assert.Equal(t, false, success)
		assert.Equal(t, "", p2pDestinationURL)
		assert.Equal(t, "", p2pSubmitTxURL)
	})

	t.Run("valid p2p capabilities", func(t *testing.T) {
		capabilities := server.GenericCapabilities(paymail.DefaultBsvAliasVersion, false)

		// Add the P2P
		capabilities.Capabilities[paymail.BRFCP2PTransactions] = "/receive-transaction/{alias}@{domain.tld}"
		capabilities.Capabilities[paymail.BRFCP2PPaymentDestination] = "/p2p-payment-destination/{alias}@{domain.tld}"

		success, p2pDestinationURL, p2pSubmitTxURL := hasP2P(capabilities)
		assert.Equal(t, true, success)
		assert.Equal(t, capabilities.Capabilities[paymail.BRFCP2PPaymentDestination], p2pDestinationURL)
		assert.Equal(t, capabilities.Capabilities[paymail.BRFCP2PTransactions], p2pSubmitTxURL)
	})
}

// Test_startP2PTransaction will test the method startP2PTransaction()
func Test_startP2PTransaction(t *testing.T) {
	// t.Parallel() mocking does not allow parallel tests

	t.Run("[mocked] - valid response", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		mockValidResponse(http.StatusOK, true, testDomain)

		payload, err := startP2PTransaction(
			client, testAlias, testDomain,
			testServerURL+"/p2p-payment-destination/{alias}@{domain.tld}", 1000,
		)
		require.NoError(t, err)
		require.NotNil(t, payload)
		assert.Equal(t, "z0bac4ec-6f15-42de-9ef4-e60bfdabf4f7", payload.Reference)
		assert.Equal(t, 1, len(payload.Outputs))
		assert.Equal(t, "16fkwYn8feXEbK7iCTg5KMx9Rx9GzZ9HuE", payload.Outputs[0].Address)
		assert.Equal(t, uint64(1000), payload.Outputs[0].Satoshis)
		assert.Equal(t, "76a9143e2d1d795f8acaa7957045cc59376177eb04a3c588ac", payload.Outputs[0].Script)
	})

	t.Run("error - address not found", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		httpmock.Reset()
		httpmock.RegisterResponder(http.MethodPost, testServerURL+"/p2p-payment-destination/"+testAlias+"@"+testDomain,
			httpmock.NewStringResponder(
				http.StatusNotFound,
				`{"message": "not found"}`,
			),
		)

		payload, err := startP2PTransaction(
			client, testAlias, testDomain,
			testServerURL+"/p2p-payment-destination/{alias}@{domain.tld}", 1000,
		)

		require.Error(t, err)
		assert.Nil(t, payload)
	})
}

// Test_getCapabilities will test the method getCapabilities()
func Test_getCapabilities(t *testing.T) {
	// t.Parallel() mocking does not allow parallel tests

	t.Run("[mocked] - valid response - no cache found", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		redisClient, redisConn := xtester.LoadMockRedis(
			testIdleTimeout,
			testMaxConnLifetime,
			testMaxActiveConnections,
			testMaxIdleConnections,
		)

		tc, err := NewClient(context.Background(),
			WithRedisConnection(redisClient),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}),
			WithChainstateOptions(false, false),
			WithDebugging(),
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer func() {
			time.Sleep(1 * time.Second)
			CloseClient(context.Background(), t, tc)
		}()

		// Get command
		getCmd := redisConn.Command(cache.GetCommand, cacheKeyCapabilities+testDomain).Expect(nil)

		mockValidResponse(http.StatusOK, false, testDomain)
		var payload *paymail.CapabilitiesPayload
		payload, err = getCapabilities(
			context.Background(), tc.Cachestore(), client, testDomain,
		)
		require.NoError(t, err)
		require.NotNil(t, payload)
		assert.Equal(t, true, getCmd.Called)
		assert.Equal(t, paymail.DefaultBsvAliasVersion, payload.BsvAlias)
		assert.Equal(t, 3, len(payload.Capabilities))
	})

	t.Run("[mocked] - server error", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		redisClient, redisConn := xtester.LoadMockRedis(
			testIdleTimeout,
			testMaxConnLifetime,
			testMaxActiveConnections,
			testMaxIdleConnections,
		)

		tc, err := NewClient(context.Background(),
			WithRedisConnection(redisClient),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}),
			WithChainstateOptions(false, false),
			WithDebugging(),
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer func() {
			time.Sleep(1 * time.Second)
			CloseClient(context.Background(), t, tc)
		}()

		// Get command
		getCmd := redisConn.Command(cache.GetCommand, cacheKeyCapabilities+testDomain).Expect(nil)

		mockValidResponse(http.StatusBadRequest, false, testDomain)
		var payload *paymail.CapabilitiesPayload
		payload, err = getCapabilities(
			context.Background(), tc.Cachestore(), client, testDomain,
		)
		require.Error(t, err)
		require.Nil(t, payload)
		assert.Equal(t, true, getCmd.Called)
	})

	t.Run("valid response - no cache found", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer func() {
			time.Sleep(1 * time.Second)
			CloseClient(context.Background(), t, tc)
		}()

		mockValidResponse(http.StatusOK, false, testDomain)
		var payload *paymail.CapabilitiesPayload
		payload, err = getCapabilities(
			context.Background(), tc.Cachestore(), client, testDomain,
		)
		require.NoError(t, err)
		require.NotNil(t, payload)
		assert.Equal(t, paymail.DefaultBsvAliasVersion, payload.BsvAlias)
		assert.Equal(t, 3, len(payload.Capabilities))
	})

	t.Run("multiple requests for same capabilities", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer func() {
			time.Sleep(1 * time.Second)
			CloseClient(context.Background(), t, tc)
		}()

		mockValidResponse(http.StatusOK, false, testDomain)
		var payload *paymail.CapabilitiesPayload
		payload, err = getCapabilities(
			context.Background(), tc.Cachestore(), client, testDomain,
		)
		require.NoError(t, err)
		require.NotNil(t, payload)
		assert.Equal(t, paymail.DefaultBsvAliasVersion, payload.BsvAlias)
		assert.Equal(t, 3, len(payload.Capabilities))

		time.Sleep(1 * time.Second)

		payload, err = getCapabilities(
			context.Background(), tc.Cachestore(), client, testDomain,
		)
		require.NoError(t, err)
		require.NotNil(t, payload)
		assert.Equal(t, paymail.DefaultBsvAliasVersion, payload.BsvAlias)
		assert.Equal(t, 3, len(payload.Capabilities))
	})
}

// Test_resolvePaymailAddress will test the method resolvePaymailAddress()
func Test_resolvePaymailAddress(t *testing.T) {
	// t.Parallel() mocking does not allow parallel tests

	t.Run("[mocked] - valid response - no cache found", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		redisClient, redisConn := xtester.LoadMockRedis(
			testIdleTimeout,
			testMaxConnLifetime,
			testMaxActiveConnections,
			testMaxIdleConnections,
		)

		tc, err := NewClient(context.Background(),
			WithRedisConnection(redisClient),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}),
			WithChainstateOptions(false, false),
			WithDebugging(),
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer func() {
			time.Sleep(1 * time.Second)
			CloseClient(context.Background(), t, tc)
		}()

		// Get command
		getCmd := redisConn.Command(cache.GetCommand, cacheKeyCapabilities+testDomain).Expect(nil)

		// Mock all responses
		mockValidResponse(http.StatusOK, false, testDomain)

		// Get capabilities
		var payload *paymail.CapabilitiesPayload
		payload, err = getCapabilities(
			context.Background(), tc.Cachestore(), client, testDomain,
		)
		require.NoError(t, err)
		require.NotNil(t, payload)
		assert.Equal(t, true, getCmd.Called)

		// Get command
		getCmd2 := redisConn.Command(cache.GetCommand, cacheKeyAddressResolution+testAlias+"-"+testDomain).Expect(nil)

		// Resolve address
		var resolvePayload *paymail.ResolutionPayload
		resolvePayload, err = resolvePaymailAddress(
			context.Background(), tc.Cachestore(), client, payload,
			testAlias, testDomain, defaultAddressResolutionPurpose, defaultSenderPaymail,
		)
		require.NoError(t, err)
		require.NotNil(t, resolvePayload)
		assert.Equal(t, true, getCmd2.Called)
		assert.Equal(t, "1Cat862cjhp8SgLLMvin5gyk5UScasg1P9", resolvePayload.Address)
		assert.Equal(t, "76a9147f11c8f67a2781df0400ebfb1f31b4c72a780b9d88ac", resolvePayload.Output)
		assert.Equal(t, "", resolvePayload.Signature)
	})

	t.Run("valid response - no cache found", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer func() {
			time.Sleep(1 * time.Second)
			CloseClient(context.Background(), t, tc)
		}()

		// Mock all responses
		mockValidResponse(http.StatusOK, false, testDomain)

		// Get capabilities
		var payload *paymail.CapabilitiesPayload
		payload, err = getCapabilities(
			context.Background(), tc.Cachestore(), client, testDomain,
		)
		require.NoError(t, err)
		require.NotNil(t, payload)

		// Resolve address
		var resolvePayload *paymail.ResolutionPayload
		resolvePayload, err = resolvePaymailAddress(
			context.Background(), tc.Cachestore(), client, payload,
			testAlias, testDomain, defaultAddressResolutionPurpose, defaultSenderPaymail,
		)
		require.NoError(t, err)
		require.NotNil(t, resolvePayload)
		assert.Equal(t, "1Cat862cjhp8SgLLMvin5gyk5UScasg1P9", resolvePayload.Address)
		assert.Equal(t, "76a9147f11c8f67a2781df0400ebfb1f31b4c72a780b9d88ac", resolvePayload.Output)
		assert.Equal(t, "", resolvePayload.Signature)
	})

	t.Run("multiple requests for same address resolution", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer func() {
			time.Sleep(1 * time.Second)
			CloseClient(context.Background(), t, tc)
		}()

		// Mock all responses
		mockValidResponse(http.StatusOK, false, testDomain)

		// Get capabilities
		var payload *paymail.CapabilitiesPayload
		payload, err = getCapabilities(
			context.Background(), tc.Cachestore(), client, testDomain,
		)
		require.NoError(t, err)
		require.NotNil(t, payload)

		// Resolve address
		var resolvePayload *paymail.ResolutionPayload
		resolvePayload, err = resolvePaymailAddress(
			context.Background(), tc.Cachestore(), client, payload,
			testAlias, testDomain, defaultAddressResolutionPurpose, defaultSenderPaymail,
		)
		require.NoError(t, err)
		require.NotNil(t, resolvePayload)
		assert.Equal(t, "1Cat862cjhp8SgLLMvin5gyk5UScasg1P9", resolvePayload.Address)

		time.Sleep(1 * time.Second)

		// Resolve address
		resolvePayload, err = resolvePaymailAddress(
			context.Background(), tc.Cachestore(), client, payload,
			testAlias, testDomain, defaultAddressResolutionPurpose, defaultSenderPaymail,
		)
		require.NoError(t, err)
		require.NotNil(t, resolvePayload)
		assert.Equal(t, "1Cat862cjhp8SgLLMvin5gyk5UScasg1P9", resolvePayload.Address)
	})
}
