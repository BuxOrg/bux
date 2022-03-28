package bux

import (
	"context"
	"database/sql"
	"time"

	"github.com/BuxOrg/bux/utils"
)

// NewDestination will get a new destination for an existing xPub
//
// xPubKey is the raw public xPub
func (c *Client) NewDestination(ctx context.Context, xPubKey string, chain uint32,
	destinationType string, monitor bool, opts ...ModelOps) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_destination")

	// Validate that the xPubKey is a valid xPub
	_, err := utils.ValidateXPub(xPubKey)
	if err != nil {
		return nil, err
	}

	// Get the xPub (by key - converts to id)
	var xPub *Xpub
	if xPub, err = getXpub(
		ctx, xPubKey, // Get the xPub by xPubID
		c.DefaultModelOptions()..., // Passing down the Datastore and client information into the model
	); err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	// Get/create a new destination
	var destination *Destination
	if destination, err = xPub.getNewDestination(
		ctx, chain, destinationType, opts...,
	); err != nil {
		return nil, err
	}

	if monitor {
		destination.Monitor = utils.NullTime{NullTime: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		}}
	}

	// Save the destination
	if err = destination.Save(ctx); err != nil {
		return nil, err
	}

	// Return the model
	return destination, nil
}

// NewDestinationForLockingScript will create a new destination based on a locking script
func (c *Client) NewDestinationForLockingScript(ctx context.Context, xPubID, lockingScript string,
	monitor bool, opts ...ModelOps) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_destination_for_locking_script")

	// Ensure locking script isn't empty
	if len(lockingScript) == 0 {
		return nil, ErrMissingLockingScript
	}

	// Start the new destination - will detect type
	destination := newDestination(
		xPubID, lockingScript,
		opts...,
	)

	if destination.Type == "" {
		return nil, ErrUnknownLockingScript
	}

	// set the monitoring, passed down from the initiating function
	// this will be set when calling NewDestination from http / graphql, but not for instance paymail
	if monitor {
		destination.Monitor = utils.NullTime{NullTime: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		}}
	}

	// Save the destination
	if err := destination.Save(ctx); err != nil {
		return nil, err
	}

	// Return the model
	return destination, nil
}

// GetDestinations will get destinations based on an xPub
//
// metadataConditions are the search criteria used to find destinations
func (c *Client) GetDestinations(ctx context.Context, xPubID string, metadataConditions *Metadata) ([]*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_destinations")

	// Get the destinations
	// todo: add params for: page size and page (right now it is unlimited)
	destinations, err := getDestinationsByXpubID(
		ctx, xPubID, metadataConditions, 0, 0, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}

	return destinations, nil
}

// GetDestinationByID will get a destination by id
func (c *Client) GetDestinationByID(ctx context.Context, xPubID, id string) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_destination_by_id")

	// Get the destinations
	destination, err := getDestinationByID(
		ctx, id, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}
	if destination == nil {
		return nil, ErrMissingDestination
	}

	// Check that the id matches
	if destination.XpubID != xPubID {
		return nil, ErrXpubIDMisMatch
	}

	return destination, nil
}

// GetDestinationByLockingScript will get a destination for a locking script
func (c *Client) GetDestinationByLockingScript(ctx context.Context, xPubID, lockingScript string) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_destination_by_locking_script")

	// Get the destinations
	destination, err := getDestinationByLockingScript(
		ctx, lockingScript, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}
	if destination == nil {
		return nil, ErrMissingDestination
	}

	// Check that the id matches
	if destination.XpubID != xPubID {
		return nil, ErrXpubIDMisMatch
	}

	return destination, nil
}

// GetDestinationByAddress will get a destination for an address
func (c *Client) GetDestinationByAddress(ctx context.Context, xPubID, address string) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_destination_by_address")

	// Get the destinations
	destination, err := getDestinationByAddress(
		ctx, address, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}
	if destination == nil {
		return nil, ErrMissingDestination
	}

	// Check that the id matches
	if destination.XpubID != xPubID {
		return nil, ErrXpubIDMisMatch
	}

	return destination, nil
}

// UpdateDestinationMetadataByID will update the metadata in an existing destination by id
func (c *Client) UpdateDestinationMetadataByID(ctx context.Context, xPubID, id string, metadata Metadata) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "update_destination_by_id")

	// Get the destinations
	destination, err := c.GetDestinationByID(ctx, xPubID, id)
	if err != nil {
		return nil, err
	}

	destination.UpdateMetadata(metadata)
	err = destination.Save(ctx)
	if err != nil {
		return nil, err
	}

	return destination, nil
}

// UpdateDestinationMetadataByLockingScript will update the metadata in an existing destination by locking script
func (c *Client) UpdateDestinationMetadataByLockingScript(ctx context.Context, xPubID, lockingScript string, metadata Metadata) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "update_destination_by_id")

	// Get the destinations
	destination, err := c.GetDestinationByLockingScript(ctx, xPubID, lockingScript)
	if err != nil {
		return nil, err
	}

	destination.UpdateMetadata(metadata)
	err = destination.Save(ctx)
	if err != nil {
		return nil, err
	}

	return destination, nil
}

// UpdateDestinationMetadataByAddress will update the metadata in an existing destination by address
func (c *Client) UpdateDestinationMetadataByAddress(ctx context.Context, xPubID, address string, metadata Metadata) (*Destination, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "update_destination_by_id")

	// Get the destinations
	destination, err := c.GetDestinationByAddress(ctx, xPubID, address)
	if err != nil {
		return nil, err
	}

	destination.UpdateMetadata(metadata)
	err = destination.Save(ctx)
	if err != nil {
		return nil, err
	}

	return destination, nil
}
