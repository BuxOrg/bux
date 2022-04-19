package chainstate

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/logger"
	"github.com/mrz1836/go-mattercloud"
	"github.com/mrz1836/go-nownodes"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-minercraft"
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

	// Set the default options
	return &clientOptions{
		config: &syncConfig{
			httpClient: nil,
			mAPI: &mAPIConfig{
				broadcastMiners: bm,
				miners:          bm,
				queryMiners:     qm,
			},
			matterCloud:       nil,
			matterCloudAPIKey: "",
			minercraft:        nil,
			network:           MainNet,
			queryTimeout:      defaultQueryTimeOut,
			whatsOnChain:      nil,
		},
		debug:           false,
		newRelicEnabled: false,
	}
}

// defaultMiners will return the miners for default configuration
func defaultMiners() (broadcastMiners []*minercraft.Miner, queryMiners []*minercraft.Miner) {

	// Set the broadcast miners
	broadcastMiners, _ = minercraft.DefaultMiners()

	// Loop and add (only miners that support ALL TX QUERY)
	for index, miner := range broadcastMiners {
		// minercraft.MinerGorillaPool, (does not have -t index enabled)
		// minercraft.MinerMatterpool, (does not have -t index enabled)
		if miner.Name == minercraft.MinerTaal || miner.Name == minercraft.MinerMempool {
			queryMiners = append(queryMiners, broadcastMiners[index])
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

// WithMempoolMonitoring will enable mempool monitoring for a given Filter
/*func WithMempoolMonitoring(handler whatsonchain.SocketHandler, Filter string) ClientOps {
	return func(c *clientOptions) {
		c.mempoolMonitoringEnabled = true
		c.mempoolMonitoringFilter = Filter
		c.mempoolHandler = handler
	}
}*/

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

// WithWhatsOnChain will set a custom WhatsOnChain client
func WithWhatsOnChain(client whatsonchain.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.config.whatsOnChain = client
		}
	}
}

// WithMatterCloud will set a custom MatterCloud client
func WithMatterCloud(client mattercloud.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.config.matterCloud = client
		}
	}
}

// WithNowNodes will set a custom NowNodes client
func WithNowNodes(client nownodes.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.config.nowNodes = client
		}
	}
}

// WithNowNodesAPIKey will set a custom NowNodes API key
func WithNowNodesAPIKey(apiKey string) ClientOps {
	return func(c *clientOptions) {
		if len(apiKey) > 0 {
			c.config.nowNodesAPIKey = apiKey
		}
	}
}

// WithMatterCloudAPIKey will set a custom MatterCloud API key
func WithMatterCloudAPIKey(apiKey string) ClientOps {
	return func(c *clientOptions) {
		if len(apiKey) > 0 {
			c.config.matterCloudAPIKey = apiKey
		}
	}
}

// WithWhatsOnChainAPIKey will set a custom WhatsOnChain API key
func WithWhatsOnChainAPIKey(apiKey string) ClientOps {
	return func(c *clientOptions) {
		if len(apiKey) > 0 {
			c.config.whatsOnChainAPIKey = apiKey
		}
	}
}

// WithBroadcastMiners will set a list of miners for broadcasting
func WithBroadcastMiners(miners []*minercraft.Miner) ClientOps {
	return func(c *clientOptions) {
		if len(miners) > 0 {
			c.config.mAPI.broadcastMiners = miners
		}
	}
}

// WithQueryMiners will set a list of miners for querying transactions
func WithQueryMiners(miners []*minercraft.Miner) ClientOps {
	return func(c *clientOptions) {
		if len(miners) > 0 {
			c.config.mAPI.queryMiners = miners
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

// WithCustomMiners will overwrite the default list of miners in Minercraft
func WithCustomMiners(miners []*minercraft.Miner) ClientOps {
	return func(c *clientOptions) {
		if c != nil && len(miners) > 0 {
			c.config.mAPI.miners = miners
		}
	}
}

// WithLogger will set a custom logger
func WithLogger(customLogger logger.Interface) ClientOps {
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
