package chainstate

import "errors"

func (c *Client) startMempoolMonitor(filter string) error {
	// start web socket - only supported for whatsonchain; error out if client doesn't exist
	if c.WhatsOnChain() == nil {
		return errors.New("whatsonchain client not configured")
	}
	return nil
}
