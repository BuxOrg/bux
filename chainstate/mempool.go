package chainstate

import (
	"errors"
)

func (c *Client) startMempoolMonitor() error {
	// start web socket - only supported for whatsonchain; error out if client doesn't exist
	if c.WhatsOnChain() == nil {
		return errors.New("whatsonchain client not configured")
	}
	c.WhatsOnChain().NewMempoolWebsocket(c.options.mempoolHandler, c.options.mempoolMonitoringFilter)
	return nil
}
