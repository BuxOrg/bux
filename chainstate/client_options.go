package chainstate

import (
	"context"
	"time"

	"github.com/bitcoin-sv/go-broadcast-client/broadcast"
	broadcastClient "github.com/bitcoin-sv/go-broadcast-client/broadcast/broadcast-client"
	zLogger "github.com/mrz1836/go-logger"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-minercraft/v2"
)

// ClientOps allow functional options to be supplied
// that overwrite default client options.
type ClientOps func(c *clientOptions)

// defaultClientOptions will return an clientOptions struct with the default settings
//
// Useful for starting with the default and then modifying as needed
func defaultClientOptions() *clientOptions {

	// Create the default miners
	bm, qm := defaultMiners()
	apis, _ := minercraft.DefaultMinersAPIs()

	// Set the default options
	return &clientOptions{
		config: &syncConfig{
			httpClient: nil,
			minercraftConfig: &minercraftConfig{
				broadcastMiners:     bm,
				queryMiners:         qm,
				minerAPIs:           apis,
				minercraftFeeQuotes: true,
				feeUnit:             DefaultFee,
			},
			minercraft:      nil,
			network:         MainNet,
			queryTimeout:    defaultQueryTimeOut,
			broadcastClient: nil,
			broadcastClientConfig: &broadcastClientConfig{
				BroadcastClientApis: nil,
			},
		},
		debug:           false,
		newRelicEnabled: false,
	}
}

// defaultMiners will return the miners for default configuration
func defaultMiners() (broadcastMiners []*Miner, queryMiners []*Miner) {
	// Set the broadcast miners
	miners, _ := minercraft.DefaultMiners()

	// Loop and add (only miners that support ALL TX QUERY)
	for index, miner := range miners {
		broadcastMiners = append(broadcastMiners, &Miner{
			FeeLastChecked: time.Now().UTC(),
			FeeUnit:        DefaultFee,
			Miner:          miners[index],
		})

		// Only miners that support querying
		if miner.Name == minercraft.MinerTaal || miner.Name == minercraft.MinerMempool {
			// minercraft.MinerGorillaPool, (does not have -t index enabled - 4.25.22)
			// minercraft.MinerMatterpool, (does not have -t index enabled - 4.25.22)
			queryMiners = append(queryMiners, &Miner{
				// FeeLastChecked: time.Now().UTC(),
				// FeeUnit:        DefaultFee,
				Miner: miners[index],
			})
		}
	}
	return
}

// getTxnCtx will check for an existing transaction
func (c *clientOptions) getTxnCtx(ctx context.Context) context.Context {
	if c.newRelicEnabled {
		txn := newrelic.FromContext(ctx)
		if txn != nil {
			ctx = newrelic.NewContext(ctx, txn)
		}
	}
	return ctx
}

// WithNewRelic will enable the NewRelic wrapper
func WithNewRelic() ClientOps {
	return func(c *clientOptions) {
		c.newRelicEnabled = true
	}
}

// WithDebugging will enable debugging mode
func WithDebugging() ClientOps {
	return func(c *clientOptions) {
		c.debug = true
	}
}

// WithHTTPClient will set a custom HTTP client
func WithHTTPClient(client HTTPInterface) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.config.httpClient = client
		}
	}
}

// WithMinercraft will set a custom Minercraft client
func WithMinercraft(client minercraft.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.config.minercraft = client
		}
	}
}

// WithMAPI will specify mAPI as an API for minercraft client
func WithMAPI() ClientOps {
	return func(c *clientOptions) {
		c.config.minercraftConfig.apiType = minercraft.MAPI
	}
}

// WithArc will specify Arc as an API for minercraft client
func WithArc() ClientOps {
	return func(c *clientOptions) {
		c.config.minercraftConfig.apiType = minercraft.Arc
	}
}

// WithBroadcastMiners will set a list of miners for broadcasting
func WithBroadcastMiners(miners []*Miner) ClientOps {
	return func(c *clientOptions) {
		if len(miners) > 0 {
			c.config.minercraftConfig.broadcastMiners = miners
		}
	}
}

// WithQueryMiners will set a list of miners for querying transactions
func WithQueryMiners(miners []*Miner) ClientOps {
	return func(c *clientOptions) {
		if len(miners) > 0 {
			c.config.minercraftConfig.queryMiners = miners
		}
	}
}

// WithQueryTimeout will set a different timeout for transaction querying
func WithQueryTimeout(timeout time.Duration) ClientOps {
	return func(c *clientOptions) {
		if timeout > 0 {
			c.config.queryTimeout = timeout
		}
	}
}

// WithUserAgent will set the custom user agent
func WithUserAgent(agent string) ClientOps {
	return func(c *clientOptions) {
		if len(agent) > 0 {
			c.userAgent = agent
		}
	}
}

// WithNetwork will set the network to use
func WithNetwork(network Network) ClientOps {
	return func(c *clientOptions) {
		if len(network) > 0 {
			c.config.network = network
		}
	}
}

// WithLogger will set a custom logger
func WithLogger(customLogger zLogger.GormLoggerInterface) ClientOps {
	return func(c *clientOptions) {
		if customLogger != nil {
			c.logger = customLogger
		}
	}
}

// WithMonitoring will create a new monitorConfig interface with the given options
func WithMonitoring(ctx context.Context, monitorOptions *MonitorOptions) ClientOps {
	return func(c *clientOptions) {
		if monitorOptions != nil {
			// Create the default Monitor for monitoring destinations
			c.monitor = NewMonitor(ctx, monitorOptions)
		}
	}
}

// WithMonitoringInterface will set the interface to use for monitoring the blockchain
func WithMonitoringInterface(monitor MonitorService) ClientOps {
	return func(c *clientOptions) {
		if monitor != nil {
			c.monitor = monitor
		}
	}
}

// WithExcludedProviders will set a list of excluded providers
func WithExcludedProviders(providers []string) ClientOps {
	return func(c *clientOptions) {
		if len(providers) > 0 {
			c.config.excludedProviders = providers
		}
	}
}

// WithMinercraftFeeQuotes will set minercraftFeeQuotes flag as true
func WithMinercraftFeeQuotes() ClientOps {
	return func(c *clientOptions) {
		c.config.minercraftConfig.minercraftFeeQuotes = true
	}
}

// WithMinercraftAPIs will set miners APIs
func WithMinercraftAPIs(apis []*minercraft.MinerAPIs) ClientOps {
	return func(c *clientOptions) {
		c.config.minercraftConfig.minerAPIs = apis
	}
}

// WithBroadcastClient will set broadcast client APIs
func WithBroadcastClient(client broadcast.Client) ClientOps {
	return func(c *clientOptions) {
		c.config.broadcastClient = client
	}
}

// WithBroadcastClientAPIs will set broadcast client APIs
func WithBroadcastClientAPIs(apis []broadcastClient.ArcClientConfig) ClientOps {
	return func(c *clientOptions) {
		c.config.broadcastClientConfig.BroadcastClientApis = apis
	}
}
