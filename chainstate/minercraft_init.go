package chainstate

import (
	"context"
	"sync"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-minercraft/v2"
	"github.com/tonicpow/go-minercraft/v2/apis/mapi"
)

func (c *Client) minercraftInit(ctx context.Context) (feeUnit *utils.FeeUnit, err error) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("start_minercraft").End()
	}
	mi := &minercraftInitializer{client: c, ctx: ctx}

	if err = mi.newClient(); err != nil {
		return
	}

	if err = mi.validateMiners(); err != nil {
		return
	}

	if c.isFeeQuotesEnabled() {
		feeUnit = mi.lowestFee()
	} else {
		feeUnit = DefaultFee()
	}

	return
}

type minercraftInitializer struct {
	client *Client
	ctx    context.Context
}

func (i *minercraftInitializer) defaultMinercraftOptions() (opts *minercraft.ClientOptions) {
	c := i.client
	opts = minercraft.DefaultClientOptions()
	if len(c.options.userAgent) > 0 {
		opts.UserAgent = c.options.userAgent
	}
	return
}

func (i *minercraftInitializer) newClient() (err error) {
	c := i.client
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
			i.defaultMinercraftOptions(),
			c.HTTPClient(),
			c.options.config.minercraftConfig.apiType,
			optionalMiners,
			c.options.config.minercraftConfig.minerAPIs,
		)
	}
	return
}

// validateMiners will check if miner is reacheble by requesting its FeeQuote
// If there was on error on FeeQuote(), the miner will be deleted from miners list
// If usage of MapiFeeQuotes is enabled and miner is reacheble, miner's fee unit will be upadeted with MAPI fee quotes
// If FeeQuote returns some quote, but fee is not presented in it, it means that miner is valid but we can't use it's feequote
func (i *minercraftInitializer) validateMiners() error {
	ctxWithCancel, cancel := context.WithTimeout(i.ctx, 5*time.Second)
	defer cancel()

	c := i.client
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
					client.options.logger.Error().Msgf("No FeeQuote response from miner %s. Reason: %s", miner.Miner.Name, err)
					miner.FeeUnit = nil
					return
				}

				fee = quote.Quote.GetFee(mapi.FeeTypeData)
				if fee == nil {
					client.options.logger.Error().Msgf("Fee is missing in %s's FeeQuote response", miner.Miner.Name)
					return
				}
				// Arc doesn't support FeeQuote right now(2023.07.21), that's why PolicyQuote is used
			} else if c.Minercraft().APIType() == minercraft.Arc {
				quote, err := c.Minercraft().PolicyQuote(ctx, miner.Miner)
				if err != nil {
					client.options.logger.Error().Msgf("No FeeQuote response from miner %s. Reason: %s", miner.Miner.Name, err)
					miner.FeeUnit = nil
					return
				}

				fee = quote.Quote.Fees[0]
			}
			if c.isFeeQuotesEnabled() {
				miner.FeeUnit = &utils.FeeUnit{
					Satoshis: fee.MiningFee.Satoshis,
					Bytes:    fee.MiningFee.Bytes,
				}
				miner.FeeLastChecked = time.Now().UTC()
			}
		}(ctxWithCancel, c, &wg, c.options.config.minercraftConfig.broadcastMiners[index])
	}
	wg.Wait()

	i.deleteUnreacheableMiners()

	switch {
	case len(c.options.config.minercraftConfig.broadcastMiners) == 0:
		return ErrMissingBroadcastMiners
	case len(c.options.config.minercraftConfig.queryMiners) == 0:
		return ErrMissingQueryMiners
	default:
		return nil
	}
}

// deleteUnreacheableMiners deletes miners which can't be reacheable from config
func (i *minercraftInitializer) deleteUnreacheableMiners() {
	c := i.client
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

// lowestFees takes the lowest fees among all miners and sets them as the feeUnit for future transactions
func (i *minercraftInitializer) lowestFee() *utils.FeeUnit {
	miners := i.client.options.config.minercraftConfig.broadcastMiners
	fees := make([]utils.FeeUnit, len(miners))
	for index, miner := range miners {
		fees[index] = *miner.FeeUnit
	}
	lowest := utils.LowestFee(fees, *DefaultFee())
	return &lowest
}
