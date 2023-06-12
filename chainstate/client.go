package chainstate

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/BuxOrg/bux/utils"
	zLogger "github.com/mrz1836/go-logger"
	"github.com/mrz1836/go-nownodes"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-minercraft"
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
		pulseAgent      PulseService                // Pulse agent for informing about the creation of new blocks
	}

	// syncConfig holds all the configuration about the different sync processes
	syncConfig struct {
		excludedProviders  []string                     // List of provider names
		httpClient         HTTPInterface                // Custom HTTP client (Minercraft, WOC)
		mAPI               *mAPIConfig                  // mAPI configuration
		minercraft         minercraft.ClientInterface   // Minercraft client
		network            Network                      // Current network (mainnet, testnet, stn)
		nowNodes           nownodes.ClientInterface     // NOWNodes client
		nowNodesAPIKey     string                       // If set, use this key
		queryTimeout       time.Duration                // Timeout for transaction query
		whatsOnChain       whatsonchain.ClientInterface // WhatsOnChain client
		whatsOnChainAPIKey string                       // If set, use this key
	}

	// mAPIConfig is specific for mAPI configuration
	mAPIConfig struct {
		broadcastMiners []*Miner // List of loaded miners for broadcasting
		queryMiners     []*Miner // List of loaded miners for querying transactions
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
	return c.options.config.mAPI.broadcastMiners
}

// QueryMiners will return the query miners
func (c *Client) QueryMiners() []*Miner {
	return c.options.config.mAPI.queryMiners
}

// Pulse will return the Pulse client
func (c *Client) Pulse() PulseService {
	return c.options.pulseAgent
}

// RefreshFeeQuotes will update all fee quotes for all broadcasting miners in mAPI
func (c *Client) RefreshFeeQuotes(ctx context.Context) error {
	// Loop all broadcast miners
	for i := range c.options.config.mAPI.broadcastMiners {

		// Skip if recently updated (quotes usually don't change that often)
		if c.options.config.mAPI.broadcastMiners[i].FeeLastChecked.After(time.Now().UTC().Add(-defaultFeeLastCheckIgnore)) {
			continue
		}

		// Get the policy quote using the miner
		quote, err := c.Minercraft().PolicyQuote(ctx, c.options.config.mAPI.broadcastMiners[i].Miner)
		if err != nil {
			if strings.Contains(err.Error(), "validation errors") { // todo: fix issue with Taal - do they require an api key?
				return nil
			}
			return err
		}

		// Get the fee and set the fee
		fee := quote.Quote.GetFee(minercraft.FeeTypeData) // todo: data for now, since it usually is more expensive (if different)
		if fee == nil {
			return errors.New("fee is missing from miner response")
		}
		c.options.config.mAPI.broadcastMiners[i].FeeUnit = &utils.FeeUnit{
			Satoshis: fee.MiningFee.Satoshis,
			Bytes:    fee.MiningFee.Bytes,
		}
		c.options.config.mAPI.broadcastMiners[i].FeeLastChecked = time.Now().UTC()
	}

	return nil
}
