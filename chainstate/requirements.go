package chainstate

// RequiredIn is the requirements for querying transaction information
type RequiredIn string

const (
	// RequiredInMempool is the transaction in mempool? (minimum requirement for a valid response)
	RequiredInMempool RequiredIn = requiredInMempool

	// RequiredOnChain is the transaction in on-chain? (minimum requirement for a valid response)
	RequiredOnChain RequiredIn = requiredOnChain
)

// ValidRequirement will return valid if the requirement is known
func (c *Client) validRequirement(requirement RequiredIn) bool {
	return requirement == RequiredOnChain || requirement == RequiredInMempool
}

// checkRequirement will check to see if the requirement has been met
func checkRequirement(requirement RequiredIn, id string, txInfo *TransactionInfo) bool {
	if requirement == RequiredInMempool { // Good response, and only has TX
		return txInfo.ID == id
	} else if requirement == RequiredOnChain { // Good response, found block hash
		if len(txInfo.BlockHash) > 0 && txInfo.Confirmations > 0 {
			return true
		}
	}
	return false
}
