package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/utils"
)

// NewAccessKey will create a new access key for the given xpub
//
// opts are options and can include "metadata"
func (c *Client) NewAccessKey(ctx context.Context, xPubKey string, opts ...ModelOps) (*AccessKey, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_access_key")

	// Validate that the value is an xPub
	_, err := utils.ValidateXPub(xPubKey)
	if err != nil {
		return nil, err
	}

	// Get the xPub (by key - converts to id)
	var xPub *Xpub
	if xPub, err = getXpub(
		ctx, xPubKey, // Pass the context and key everytime (for now)
		c.DefaultModelOptions()..., // Passing down the Datastore and client information into the model
	); err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	// Create the model & set the default options (gives options from client->model)
	accessKey := newAccessKey(
		xPub.ID, c.DefaultModelOptions(append(opts, New())...)...,
	)

	// Save the model
	if err = accessKey.Save(ctx); err != nil {
		return nil, err
	}

	// Return the created model
	return accessKey, nil
}

// RevokeAccessKey will revoke an access key
//
// opts are options and can include "metadata"
func (c *Client) RevokeAccessKey(ctx context.Context, xPubKey, id string, opts ...ModelOps) (*AccessKey, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_access_key")

	// Validate that the value is an xPub
	_, err := utils.ValidateXPub(xPubKey)
	if err != nil {
		return nil, err
	}

	// Get the xPub (by key - converts to id)
	var xPub *Xpub
	if xPub, err = getXpub(
		ctx, xPubKey, // Pass the context and key everytime (for now)
		c.DefaultModelOptions()..., // Passing down the Datastore and client information into the model
	); err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	var accessKey *AccessKey
	if accessKey, err = GetAccessKey(
		ctx, id, c.DefaultModelOptions(opts...)...,
	); err != nil {
		return nil, err
	}

	// make sure this is the correct xPub
	if accessKey.XpubID != utils.Hash(xPubKey) {
		return nil, utils.ErrXpubNoMatch
	}

	accessKey.RevokedAt.Valid = true
	accessKey.RevokedAt.Time = time.Now()

	// Save the model
	if err = accessKey.Save(ctx); err != nil {
		return nil, err
	}

	// Return the updated model
	return accessKey, nil
}
