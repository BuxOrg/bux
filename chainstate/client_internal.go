package chainstate

import (
	"context"

	"github.com/BuxOrg/bux/utils"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-minercraft/v2"
)

// defaultMinercraftOptions will create the defaults
func (c *Client) defaultMinercraftOptions() (opts *minercraft.ClientOptions) {
	opts = minercraft.DefaultClientOptions()
	if len(c.options.userAgent) > 0 {
		opts.UserAgent = c.options.userAgent
	}
	return
}

// startMinerCraft will start Minercraft (if no custom client is found)
func (c *Client) startMinerCraft(ctx context.Context) (err error) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("start_minercraft").End()
	}

	// No client?
	if c.Minercraft() == nil {
		var optionalMiners []*minercraft.Miner
		var loadedMiners []string

		// Loop all broadcast miners and append to the list of miners
		for i := range c.options.config.minercraftConfig.broadcastMiners {
			if !utils.StringInSlice(c.options.config.minercraftConfig.broadcastMiners[i].Miner.MinerID, loadedMiners) {
				optionalMiners = append(optionalMiners, c.options.config.minercraftConfig.broadcastMiners[i].Miner)
				loadedMiners = append(loadedMiners, c.options.config.minercraftConfig.broadcastMiners[i].Miner.MinerID)
			}
		}

		// Loop all query miners and append to the list of miners
		for i := range c.options.config.minercraftConfig.queryMiners {
			if !utils.StringInSlice(c.options.config.minercraftConfig.queryMiners[i].Miner.MinerID, loadedMiners) {
				optionalMiners = append(optionalMiners, c.options.config.minercraftConfig.queryMiners[i].Miner)
				loadedMiners = append(loadedMiners, c.options.config.minercraftConfig.queryMiners[i].Miner.MinerID)
			}
		}
		c.options.config.minercraft, err = minercraft.NewClient(
			c.defaultMinercraftOptions(),
			c.HTTPClient(),
			c.options.config.minercraftConfig.apiType,
			optionalMiners,
			c.options.config.minercraftConfig.minerAPIs,
		)
	}

	c.ValidateMiners(ctx)

	// Check for broadcast miners
	if len(c.BroadcastMiners()) == 0 {
		return ErrMissingBroadcastMiners
	}

	// Check for query miners
	if len(c.QueryMiners()) == 0 {
		return ErrMissingQueryMiners
	}

	return nil
}
