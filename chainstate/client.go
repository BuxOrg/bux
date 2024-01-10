package chainstate

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/logging"
	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoin-sv/go-broadcast-client/broadcast"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"github.com/tonicpow/go-minercraft/v2"
)

type (

	// Client is the client (configuration)
	Client struct {
		options *clientOptions
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
		config          *syncConfig     // Configuration for broadcasting and other chain-state actions
		debug           bool            // For extra logs and additional debug information
		logger          *zerolog.Logger // Logger interface
		monitor         MonitorService  // Monitor service
		newRelicEnabled bool            // If NewRelic is enabled (parent application)
		userAgent       string          // Custom user agent for outgoing HTTP Requests
	}

	// syncConfig holds all the configuration about the different sync processes
	syncConfig struct {
		excludedProviders []string                   // List of provider names
		httpClient        HTTPInterface              // Custom HTTP client (Minercraft, WOC)
		minercraftConfig  *minercraftConfig          // minercraftConfig configuration
		minercraft        minercraft.ClientInterface // Minercraft client
		network           Network                    // Current network (mainnet, testnet, stn)
		queryTimeout      time.Duration              // Timeout for transaction query
		broadcastClient   broadcast.Client           // Broadcast client
		pulseClient       *PulseClient               // Pulse client
		feeUnit           *utils.FeeUnit             // The lowest fees among all miners
		feeQuotes         bool                       // If set, feeUnit will be updated with fee quotes from miner's
	}

	// minercraftConfig is specific for minercraft configuration
	minercraftConfig struct {
		broadcastMiners []*Miner // List of loaded miners for broadcasting
		queryMiners     []*Miner // List of loaded miners for querying transactions

		apiType   minercraft.APIType      // MinerCraft APIType(ARC/mAPI)
		minerAPIs []*minercraft.MinerAPIs // List of miners APIs
	}

	// Miner is the internal chainstate miner (wraps Minercraft miner with more information)
	Miner struct {
		FeeLastChecked time.Time         `json:"fee_last_checked"` // Last time the fee was checked via mAPI
		FeeUnit        *utils.FeeUnit    `json:"fee_unit"`         // The fee unit returned from Policy request
		Miner          *minercraft.Miner `json:"miner"`            // The minercraft miner
	}

	// PulseClient is the internal chainstate pulse client
	PulseClient struct {
		url       string
		authToken string
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
		client.options.logger = logging.GetDefaultLogger()
	}

	// Init active provider
	var feeUnit *utils.FeeUnit
	var err error
	switch client.ActiveProvider() {
	case ProviderMinercraft:
		feeUnit, err = client.minercraftInit(ctx)
	case ProviderBroadcastClient:
		feeUnit, err = client.broadcastClientInit(ctx)
	}

	if err != nil {
		return nil, err
	}

	// Set fee unit
	if feeUnit == nil {
		feeUnit = DefaultFee()
		client.options.logger.Info().Msgf("no fee unit found, using default: %s", feeUnit)
	} else {
		client.options.logger.Info().Msgf("using fee unit: %s", feeUnit)
	}
	client.options.config.feeUnit = feeUnit

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
	c.options.logger.Debug().Msg(text)
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

// BroadcastClient will return the BroadcastClient client
func (c *Client) BroadcastClient() broadcast.Client {
	return c.options.config.broadcastClient
}

// QueryTimeout will return the query timeout
func (c *Client) QueryTimeout() time.Duration {
	return c.options.config.queryTimeout
}

// FeeUnit will return feeUnit
func (c *Client) FeeUnit() *utils.FeeUnit {
	return c.options.config.feeUnit
}

func (c *Client) isFeeQuotesEnabled() bool {
	return c.options.config.feeQuotes
}

// ActiveProvider returns a name of a provider based on config.
func (c *Client) ActiveProvider() string {
	excluded := c.options.config.excludedProviders
	if !utils.StringInSlice(ProviderBroadcastClient, excluded) && c.BroadcastClient() != nil {
		return ProviderBroadcastClient
	}
	if !utils.StringInSlice(ProviderMinercraft, excluded) && (c.Network() == MainNet || c.Network() == TestNet) {
		return ProviderMinercraft
	}
	return ProviderNone
}
