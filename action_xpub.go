package bux

import (
	"context"

	"github.com/BuxOrg/bux/utils"
)

// NewXpub will parse the xPub and Save it into the Datastore
//
// xPubKey is the raw public xPub
// opts are options and can include "metadata"
func (c *Client) NewXpub(ctx context.Context, xPubKey string, opts ...ModelOps) (*Xpub, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_xpub")

	// Create the model & set the default options (gives options from client->model)
	xPub := newXpub(
		xPubKey, c.DefaultModelOptions(append(opts, New())...)...,
	)

	// Save the model
	if err := xPub.Save(ctx); err != nil {
		return nil, err
	}

	// Return the created model
	return xPub, nil
}

// GetXpub will get an existing xPub from the Datastore
//
// xPubKey is the raw public xPub
func (c *Client) GetXpub(ctx context.Context, xPubKey string) (*Xpub, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_xpub")

	// Validate the xPub
	if _, err := utils.ValidateXPub(xPubKey); err != nil {
		return nil, err
	}

	// Get the xPub (by key - converts to id)
	xPub, err := getXpub(
		ctx, xPubKey, // Pass the context and key everytime (for now)
		c.DefaultModelOptions()..., // Passing down the Datastore and client information into the model
	)
	if err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	// Return the model
	return xPub, nil
}

// GetXpubByID will get an existing xPub from the Datastore
func (c *Client) GetXpubByID(ctx context.Context, xPubID string) (*Xpub, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_xpub_by_id")

	// Get the xPub (by key - converts to id)
	xPub, err := getXpubByID(
		ctx, xPubID,
		c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	// Return the model
	return xPub, nil
}
