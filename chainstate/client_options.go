package chainstate

import (
	"context"
	"reflect"
	"strings"
	"time"

	zLogger "github.com/mrz1836/go-logger"
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
				queryMiners:     qm,
				feeUnit:         DefaultFee,
			},
			minercraft:   nil,
			network:      MainNet,
			queryTimeout: defaultQueryTimeOut,
			whatsOnChain: nil,
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

// WithWhatsOnChain will set a custom WhatsOnChain client
func WithWhatsOnChain(client whatsonchain.ClientInterface) ClientOps {
	return func(c *clientOptions) {
		if client != nil {
			c.config.whatsOnChain = client
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

// WithWhatsOnChainAPIKey will set a custom WhatsOnChain API key
func WithWhatsOnChainAPIKey(apiKey string) ClientOps {
	return func(c *clientOptions) {
		if len(apiKey) > 0 {
			c.config.whatsOnChainAPIKey = apiKey
		}
	}
}

// WithBroadcastMiners will set a list of miners for broadcasting
func WithBroadcastMiners(miners []*Miner) ClientOps {
	return func(c *clientOptions) {
		if len(miners) > 0 {
			c.config.mAPI.broadcastMiners = miners
		}
	}
}

// WithQueryMiners will set a list of miners for querying transactions
func WithQueryMiners(miners []*Miner) ClientOps {
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

// WithMapiFeeQuotes will set mapiFeeQuotesEnabled flag as true
func WithMapiFeeQuotes() ClientOps {
	return func(c *clientOptions) {
		c.config.mAPI.mapiFeeQuotesEnabled = true
	}
}

// WithOverridenMAPIConfig will override default config
func WithOverridenMAPIConfig(miners []*minercraft.Miner) ClientOps {
	return func(c *clientOptions) {
		overrideMAPIConfig(c.config.mAPI.broadcastMiners, miners)
		overrideMAPIConfig(c.config.mAPI.queryMiners, miners)
	}
}

// Looks for miners by name over the mAPI config, and rewrites fields presented in a custom config
func overrideMAPIConfig(configToOverride []*Miner, customConfig []*minercraft.Miner) {
	for _, miner := range customConfig {
		var minerToOverride *minercraft.Miner
		for _, m := range configToOverride {
			if strings.EqualFold(m.Miner.Name, miner.Name) {
				minerToOverride = m.Miner
				break
			}
		}
		// The miner is not in the configuration, and therefore there is nothing to override. Skip
		if minerToOverride == nil {
			continue
		}
		// Reflect values of miners. Needed to loop over all miner's fields and overwrite only some of them
		// Miners are pointers in both configs, so we use reflect.ValueOf(miner).Elem()
		minerToOverrideReflect := reflect.ValueOf(minerToOverride).Elem()
		overrideReflect := reflect.ValueOf(miner).Elem()
		// We don't override 'Name' field. Should be skipped
		fieldToIgnore := overrideReflect.FieldByName("Name")

		for i := 0; i < overrideReflect.NumField(); i++ {
			newField := overrideReflect.Field(i)
			if newField == fieldToIgnore {
				continue
			}
			// Only non-zero fields from custom config will used as overwrite fields
			if !newField.IsZero() {
				name := overrideReflect.Type().Field(i).Name
				fieldToOverride := minerToOverrideReflect.FieldByName(name)
				if fieldToOverride.CanSet() {
					fieldToOverride.Set(newField)
				}
			}
		}
	}
}
