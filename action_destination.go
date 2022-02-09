package bux

import (
	"context"

	"github.com/BuxOrg/bux/utils"
)

// NewDestination will get a new destination for an existing xPub
//
// xPubKey is the raw public xPub
func (c *Client) NewDestination(ctx context.Context, xPubKey string, chain uint32,
	destinationType string, metadata *map[string]interface{}) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_destination")

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

	// Get/create a new destination
	var destination *Destination
	if destination, err = xPub.getNewDestination(
		ctx, chain, destinationType, metadata,
	); err != nil {
		return nil, err
	}

	// Save the destination
	if err = destination.Save(ctx); err != nil {
		return nil, err
	}

	// Return the model
	return destination, nil
}

// NewDestinationForLockingScript will create a new destination based on a locking script
func (c *Client) NewDestinationForLockingScript(ctx context.Context, xPubKey, lockingScript, destinationType string,
	metadata map[string]interface{}) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_destination_for_locking_script")

	// Ensure locking script isn't empty
	if len(lockingScript) == 0 {
		return nil, ErrMissingLockingScript
	}

	/*
		// Validate that the value is an xPub
		_, err := utils.ValidateXPub(xPubKey)
		if err != nil {
			return nil, err
		}
	*/

	// Start the new destination
	destination := newDestination(
		utils.Hash(xPubKey), lockingScript,
		append(c.DefaultModelOptions(), WithMetadatas(metadata))...,
	)

	// Modify destination type
	// todo: this should be built into newDestination()
	destination.Type = destinationType

	// Save the destination
	if err := destination.Save(ctx); err != nil {
		return nil, err
	}

	// Return the model
	return destination, nil
}

// GetDestinations will get destinations based on an xPub
func (c *Client) GetDestinations(ctx context.Context, xPubKey string, usingMetadata *Metadata) ([]*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_destinations")

	// Get the destinations
	// todo: add params for: page size and page (right now it is unlimited)
	destinations, err := getDestinationsByXpubID(
		ctx, utils.Hash(xPubKey), usingMetadata, 0, 0, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}

	return destinations, nil
}

// GetDestinationByLockingScript will get a destination for a locking script
func (c *Client) GetDestinationByLockingScript(ctx context.Context, xPubKey, lockingScript string) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_destination_by_locking_script")

	// Get the destinations
	destination, err := getDestinationByLockingScript(
		ctx, lockingScript, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}

	// Check that the id matches
	if destination.XpubID != utils.Hash(xPubKey) {
		return nil, ErrXpubIDMisMatch
	}

	return destination, nil
}

// GetDestinationByAddress will get a destination for an address
func (c *Client) GetDestinationByAddress(ctx context.Context, xPubKey, address string) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_destination_by_address")

	// Get the destinations
	destination, err := getDestinationByAddress(
		ctx, address, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}

	// Check that the id matches
	if destination.XpubID != utils.Hash(xPubKey) {
		return nil, ErrXpubIDMisMatch
	}

	return destination, nil
}
