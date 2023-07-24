package chainstate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
	zLogger "github.com/mrz1836/go-logger"
	"github.com/mrz1836/go-nownodes"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-minercraft/v2"
	"github.com/tonicpow/go-minercraft/v2/apis/mapi"
)

type (

	// Client is the client (configuration)
	Client struct {
		options *clientOptions
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
		config          *syncConfig                 // Configuration for broadcasting and other chain-state actions
		debug           bool                        // For extra logs and additional debug information
		logger          zLogger.GormLoggerInterface // Internal logger interface
		monitor         MonitorService              // Monitor service
		newRelicEnabled bool                        // If NewRelic is enabled (parent application)
		userAgent       string                      // Custom user agent for outgoing HTTP Requests
	}

	// syncConfig holds all the configuration about the different sync processes
	syncConfig struct {
		excludedProviders  []string                     // List of provider names
		httpClient         HTTPInterface                // Custom HTTP client (Minercraft, WOC)
		minercraftConfig   *minercraftConfig            // minercraftConfig configuration
		minercraft         minercraft.ClientInterface   // Minercraft client
		network            Network                      // Current network (mainnet, testnet, stn)
		nowNodes           nownodes.ClientInterface     // NOWNodes client
		nowNodesAPIKey     string                       // If set, use this key
		queryTimeout       time.Duration                // Timeout for transaction query
		whatsOnChain       whatsonchain.ClientInterface // WhatsOnChain client
		whatsOnChainAPIKey string                       // If set, use this key
	}

	// minercraftConfig is specific for minercraft configuration
	minercraftConfig struct {
		broadcastMiners     []*Miner                // List of loaded miners for broadcasting
		queryMiners         []*Miner                // List of loaded miners for querying transactions
		feeUnit             *utils.FeeUnit          // The lowest fees among all miners
		minercraftFeeQuotes bool                    // If set, feeUnit will be updated with fee quotes from miner's mAPI
		apiType             minercraft.APIType      // MinerCraft APIType(ARC/mAPI)
		minerAPIs           []*minercraft.MinerAPIs // List of miners APIs
	}

	// Miner is the internal chainstate miner (wraps Minercraft miner with more information)
	Miner struct {
		FeeLastChecked time.Time         `json:"fee_last_checked"` // Last time the fee was checked via mAPI
		FeeUnit        *utils.FeeUnit    `json:"fee_unit"`         // The fee unit returned from Policy request
		Miner          *minercraft.Miner `json:"miner"`            // The minercraft miner
	}
)

// NewClient creates a new client for all on-chain functionality
//
// If no options are given, it will use the defaultClientOptions()
// ctx may contain a NewRelic txn (or one will be created)
func NewClient(ctx context.Context, opts ...ClientOps) (ClientInterface, error) {

	// Create a new client with defaults
	client := &Client{options: defaultClientOptions()}

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(client.options)
	}

	// Use NewRelic if it's enabled (use existing txn if found on ctx)
	ctx = client.options.getTxnCtx(ctx)

	// Set logger if not set
	if client.options.logger == nil {
		client.options.logger = zLogger.NewGormLogger(client.IsDebug(), 4)
	}

	// Start Minercraft
	if err := client.startMinerCraft(ctx); err != nil {
		return nil, err
	}

	// Start WhatsOnChain
	client.startWhatsOnChain(ctx)

	// Start NowNodes
	client.startNowNodes(ctx)

	// Return the client
	return client, nil
}

// Close will close the client and any open connections
func (c *Client) Close(ctx context.Context) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("close_chainstate").End()
	}
	if c != nil && c.options.config != nil {

		// Close minercraft
		if c.options.config.minercraft != nil {
			c.options.config.minercraft = nil
		}

		// Close WhatsOnChain
		if c.options.config.whatsOnChain != nil {
			c.options.config.whatsOnChain = nil
		}

		// Close NowNodes
		if c.options.config.nowNodes != nil {
			c.options.config.nowNodes = nil
		}

		// Stop the active Monitor (if not already stopped)
		if c.options.monitor != nil {
			_ = c.options.monitor.Stop(ctx)
			c.options.monitor = nil
		}
	}
}

// Debug will set the debug flag
func (c *Client) Debug(on bool) {
	c.options.debug = on
}

// DebugLog will display verbose logs
func (c *Client) DebugLog(text string) {
	if c.IsDebug() {
		c.options.logger.Info(context.Background(), text)
	}
}

// IsDebug will return if debugging is enabled
func (c *Client) IsDebug() bool {
	return c.options.debug
}

// IsNewRelicEnabled will return if new relic is enabled
func (c *Client) IsNewRelicEnabled() bool {
	return c.options.newRelicEnabled
}

// HTTPClient will return the HTTP client
func (c *Client) HTTPClient() HTTPInterface {
	return c.options.config.httpClient
}

// Network will return the current network
func (c *Client) Network() Network {
	return c.options.config.network
}

// Minercraft will return the Minercraft client
func (c *Client) Minercraft() minercraft.ClientInterface {
	return c.options.config.minercraft
}

// Monitor will return the Monitor client
func (c *Client) Monitor() MonitorService {
	return c.options.monitor
}

// WhatsOnChain will return the WhatsOnChain client
func (c *Client) WhatsOnChain() whatsonchain.ClientInterface {
	return c.options.config.whatsOnChain
}

// NowNodes will return the NowNodes client
func (c *Client) NowNodes() nownodes.ClientInterface {
	return c.options.config.nowNodes
}

// QueryTimeout will return the query timeout
func (c *Client) QueryTimeout() time.Duration {
	return c.options.config.queryTimeout
}

// BroadcastMiners will return the broadcast miners
func (c *Client) BroadcastMiners() []*Miner {
	return c.options.config.minercraftConfig.broadcastMiners
}

// QueryMiners will return the query miners
func (c *Client) QueryMiners() []*Miner {
	return c.options.config.minercraftConfig.queryMiners
}

// FeeUnit will return feeUnit
func (c *Client) FeeUnit() *utils.FeeUnit {
	return c.options.config.minercraftConfig.feeUnit
}

func (c *Client) isMinercraftFeeQuotesEnabled() bool {
	return c.options.config.minercraftConfig.minercraftFeeQuotes
}

// ValidateMiners will check if miner is reacheble by requesting its FeeQuote
// If there was on error on FeeQuote(), the miner will be deleted from miners list
// If usage of MapiFeeQuotes is enabled and miner is reacheble, miner's fee unit will be upadeted with MAPI fee quotes
// If FeeQuote returns some quote, but fee is not presented in it, it means that miner is valid but we can't use it's feequote
func (c *Client) ValidateMiners(ctx context.Context) {
	ctxWithCancel, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	// Loop all broadcast miners
	for index := range c.options.config.minercraftConfig.broadcastMiners {
		wg.Add(1)
		go func(
			ctx context.Context, client *Client,
			wg *sync.WaitGroup, miner *Miner,
		) {
			defer wg.Done()
			// Get the fee quote using the miner
			// Switched from policyQuote to feeQuote as gorillapool doesn't have such endpoint
			var fee *bt.Fee
			if c.Minercraft().APIType() == minercraft.MAPI {
				quote, err := c.Minercraft().FeeQuote(ctx, miner.Miner)
				if err != nil {
					client.options.logger.Error(ctx, fmt.Sprintf("No FeeQuote response from miner %s", miner.Miner.Name))
					miner.FeeUnit = nil
					return
				}

				fee = quote.Quote.GetFee(mapi.FeeTypeData)
				if fee == nil {
					client.options.logger.Error(ctx, fmt.Sprintf("Fee is missing in %s's FeeQuote response", miner.Miner.Name))
					return
				}
				// Arc doesn't support FeeQuote right now(2023.07.21), that's why PolicyQuote is used
			} else if c.Minercraft().APIType() == minercraft.Arc {
				quote, err := c.Minercraft().PolicyQuote(ctx, miner.Miner)
				if err != nil {
					client.options.logger.Error(ctx, fmt.Sprintf("No FeeQuote response from miner %s", miner.Miner.Name))
					miner.FeeUnit = nil
					return
				}

				fee = quote.Quote.Fees[0]
			}
			if c.isMinercraftFeeQuotesEnabled() {
				miner.FeeUnit = &utils.FeeUnit{
					Satoshis: fee.MiningFee.Satoshis,
					Bytes:    fee.MiningFee.Bytes,
				}
				miner.FeeLastChecked = time.Now().UTC()
			}
		}(ctxWithCancel, c, &wg, c.options.config.minercraftConfig.broadcastMiners[index])
	}
	wg.Wait()

	c.DeleteUnreacheableMiners()

	if c.isMinercraftFeeQuotesEnabled() {
		c.SetLowestFees()
	}
}

// SetLowestFees takes the lowest fees among all miners and sets them as the feeUnit for future transactions
func (c *Client) SetLowestFees() {
	minFees := DefaultFee
	for _, m := range c.options.config.minercraftConfig.broadcastMiners {
		if float64(minFees.Satoshis)/float64(minFees.Bytes) > float64(m.FeeUnit.Satoshis)/float64(m.FeeUnit.Bytes) {
			minFees = m.FeeUnit
		}
	}
	c.options.config.minercraftConfig.feeUnit = minFees
}

// DeleteUnreacheableMiners deletes miners which can't be reacheable from config
func (c *Client) DeleteUnreacheableMiners() {
	validMinerIndex := 0
	for _, miner := range c.options.config.minercraftConfig.broadcastMiners {
		if miner.FeeUnit != nil {
			c.options.config.minercraftConfig.broadcastMiners[validMinerIndex] = miner
			validMinerIndex++
		}
	}
	// Prevent memory leak by erasing truncated miners
	for i := validMinerIndex; i < len(c.options.config.minercraftConfig.broadcastMiners); i++ {
		c.options.config.minercraftConfig.broadcastMiners[i] = nil
	}
	c.options.config.minercraftConfig.broadcastMiners = c.options.config.minercraftConfig.broadcastMiners[:validMinerIndex]
}
