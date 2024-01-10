package chainstate

import (
	"context"
	"errors"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoin-sv/go-broadcast-client/broadcast"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func (c *Client) broadcastClientInit(ctx context.Context) (feeUnit *utils.FeeUnit, err error) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("start_broadcast_client").End()
	}

	bc := c.options.config.broadcastClient
	if bc == nil {
		err = errors.New("broadcast client is not configured")
		return
	}

	feeUnit = DefaultFee()
	if c.isFeeQuotesEnabled() {
		// get the lowest fee
		var feeQuotes []*broadcast.FeeQuote
		feeQuotes, err = bc.GetFeeQuote(ctx)
		if err != nil {
			return
		}
		if len(feeQuotes) == 0 {
			c.options.logger.Warn().Msg("no fee quotes returned from broadcast client")
		}
		fees := make([]utils.FeeUnit, len(feeQuotes))
		for index, fee := range feeQuotes {
			fees[index] = utils.FeeUnit{
				Satoshis: int(fee.MiningFee.Satoshis),
				Bytes:    int(fee.MiningFee.Bytes),
			}
		}
		lowest := utils.LowestFee(fees, *DefaultFee())
		feeUnit = &lowest
	}

	return
}
