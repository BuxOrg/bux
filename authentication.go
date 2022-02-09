package bux

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2/bscript"
)

// AuthenticateRequest will parse the incoming request for the associated authentication header,
// and it will check the Key/Signature
//
// Sets req.Context(xpub) and req.Context(xpub_hash)
func (c *Client) AuthenticateRequest(ctx context.Context, req *http.Request, adminXPubs []string,
	adminRequired, requireSigning, signingDisabled bool) (*http.Request, error) {

	// Get the xPub/Access Key from the header
	xPubOrAccessKey := strings.TrimSpace(req.Header.Get(AuthHeader))
	if len(xPubOrAccessKey) == 0 { // No value found
		return req, ErrMissingAuthHeader
	}

	// Check for admin key
	if adminRequired {
		if !utils.StringInSlice(xPubOrAccessKey, adminXPubs) {
			return req, ErrNotAdminKey
		}
	}

	// adminRequired will always force checking of a signature
	if (requireSigning || adminRequired) && !signingDisabled {
		if req.Body == nil {
			return req, ErrMissingBody
		}
		defer func() {
			_ = req.Body.Close()
		}()
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return req, err
		}

		req.Body = io.NopCloser(bytes.NewReader(b))

		authTime, _ := strconv.Atoi(req.Header.Get(AuthHeaderTime))
		authData := &AuthPayload{
			AuthHash:     req.Header.Get(AuthHeaderHash),
			AuthNonce:    req.Header.Get(AuthHeaderNonce),
			AuthTime:     int64(authTime),
			BodyContents: string(b),
			Signature:    req.Header.Get(AuthSignature),
		}
		if err = c.checkSignature(ctx, xPubOrAccessKey, authData); err != nil {
			return req, err
		}
	} else {
		// Validate that the xPub is an HD key (length, validation)
		// NOTE: you can not use an access key if signing is turned off
		if _, err := utils.ValidateXPub(xPubOrAccessKey); err != nil {
			return req, err
		}
	}

	// Set the data back onto the request
	return setOnRequest(setOnRequest(req, xPubKey, xPubOrAccessKey), xPubHashKey, utils.Hash(xPubOrAccessKey)), nil
}

// checkSignature check the signature for the provided auth payload
func (c *Client) checkSignature(ctx context.Context, xPubOrAccessKey string, auth *AuthPayload) error {

	// Check that we have the basic signature components
	if err := checkSignatureRequirements(auth); err != nil {
		return err
	}

	// Check xPub vs Access Key
	if strings.Contains(xPubOrAccessKey, "xpub") && len(xPubOrAccessKey) > 64 {
		return verifyKeyXPub(xPubOrAccessKey, auth)
	}
	return verifyAccessKey(ctx, xPubOrAccessKey, auth, c.DefaultModelOptions()...)
}

// checkSignatureRequirements will check the payload for basic signature requirements
func checkSignatureRequirements(auth *AuthPayload) error {

	// Check that we have a signature
	if auth == nil || auth.Signature == "" {
		return ErrMissingSignature
	}

	// Check the auth hash vs the body hash
	if auth.AuthHash != createBodyHash(auth.BodyContents) {
		return ErrAuhHashMismatch
	}

	// Check the auth timestamp
	if time.Now().UTC().After(time.UnixMilli(auth.AuthTime).Add(AuthSignatureTTL)) {
		return ErrSignatureExpired
	}
	return nil
}

// verifyKeyXPub will verify the xPub key and the signature payload
func verifyKeyXPub(xPub string, auth *AuthPayload) error {

	// Validate that the xPub is an HD key (length, validation)
	if _, err := utils.ValidateXPub(xPub); err != nil {
		return err
	}

	// Cannot be nil
	if auth == nil {
		return ErrMissingSignature
	}

	// Get the key from xPub
	key, err := bitcoin.GetHDKeyFromExtendedPublicKey(xPub)
	if err != nil {
		return err
	}

	// Derive the address for signing
	if key, err = utils.DeriveChildKeyFromHex(key, auth.AuthNonce); err != nil {
		return err
	}

	var address *bscript.Address
	if address, err = bitcoin.GetAddressFromHDKey(key); err != nil {
		return err // Should never error
	}

	// Return the error if verification fails
	if err = bitcoin.VerifyMessage(
		address.AddressString,
		auth.Signature,
		getSigningMessage(xPub, auth),
	); err != nil {
		return ErrSignatureInvalid
	}
	return nil
}

// verifyAccessKey will verify the access key and the signature payload
func verifyAccessKey(ctx context.Context, key string, auth *AuthPayload, opts ...ModelOps) error {

	// Get access key from DB
	// todo: add caching in the future, faster than DB
	accessKey, err := GetAccessKey(ctx, utils.Hash(key), opts...)
	if err != nil {
		return err
	} else if accessKey == nil {
		return ErrUnknownAccessKey
	} else if accessKey.RevokedAt.Valid {
		return ErrAccessKeyRevoked
	}

	// todo: should this be the access key or the xpub id?
	var address *bscript.Address
	if address, err = bitcoin.GetAddressFromPubKeyString(
		key, true,
	); err != nil {
		return err
	}

	// Return the error if verification fails
	if err = bitcoin.VerifyMessage(
		address.AddressString,
		auth.Signature,
		getSigningMessage(key, auth),
	); err != nil {
		return ErrSignatureInvalid
	}
	return nil
}

// SetSignature will set the signature on the header for the request
func SetSignature(header *http.Header, xPriv *bip32.ExtendedKey, bodyString string) error {

	// Create the signature
	authData, err := createSignature(xPriv, bodyString)
	if err != nil {
		return err
	}

	// Set the auth header
	header.Set(AuthHeader, authData.xPub)

	// Create the auth header hash
	header.Set(AuthHeaderHash, authData.AuthHash)

	// Set the nonce
	header.Set(AuthHeaderNonce, authData.AuthNonce)

	// Set the time
	header.Set(AuthHeaderTime, fmt.Sprintf("%d", authData.AuthTime))

	// Set the signature
	header.Set(AuthSignature, authData.Signature)

	return nil
}

// CreateSignature will create a signature for the given key & body contents
func CreateSignature(xPriv *bip32.ExtendedKey, bodyString string) (string, error) {
	authData, err := createSignature(xPriv, bodyString)
	if err != nil {
		return "", err
	}
	return authData.Signature, nil
}

// getSigningMessage will build the signing message string
func getSigningMessage(xPub string, auth *AuthPayload) string {
	return fmt.Sprintf("%s%s%s%d", xPub, auth.AuthHash, auth.AuthNonce, auth.AuthTime)
}

// GetXpubFromRequest gets the stored xPub from the request if found
func GetXpubFromRequest(req *http.Request) (string, bool) {
	return getFromRequest(req, xPubKey)
}

// GetXpubHashFromRequest gets the stored xPub hash from the request if found
func GetXpubHashFromRequest(req *http.Request) (string, bool) {
	return getFromRequest(req, xPubHashKey)
}
