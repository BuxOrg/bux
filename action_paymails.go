package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/utils"
)

// NewPaymailAddress will create a new paymail address
func (c *Client) NewPaymailAddress(ctx context.Context, xPubKey, address string, opts ...ModelOps) (*PaymailAddress, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_paymail_address")

	// Start the new paymail address model
	paymailAddress := newPaymail(address, append(opts, New(), WithXPub(xPubKey))...)

	// Save the model
	if err := paymailAddress.Save(ctx); err != nil {
		return nil, err
	}

	return paymailAddress, nil
}

// DeletePaymailAddress will delete a paymail address
func (c *Client) DeletePaymailAddress(ctx context.Context, address string, opts ...ModelOps) error {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "delete_paymail_address")

	// Get the paymail address
	paymailAddress, err := getPaymail(ctx, address, opts...)
	if err != nil {
		return err
	} else if paymailAddress == nil {
		return ErrMissingPaymail
	}

	// todo: make a better approach for deleting paymail addresses?
	var randomString string
	if randomString, err = utils.RandomHex(16); err != nil {
		return err
	}

	// We will do a soft delete to make sure we still have the history for this address
	// setting the Domain to a random string solved the problem of the unique index on Alias/Domain
	paymailAddress.Alias = paymailAddress.Alias + "@" + paymailAddress.Domain
	paymailAddress.Domain = randomString
	paymailAddress.DeletedAt.Valid = true
	paymailAddress.DeletedAt.Time = time.Now()

	return paymailAddress.Save(ctx)
}
