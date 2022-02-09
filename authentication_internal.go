package bux

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
)

const (
	// AuthHeader is the header to use for authentication (raw xPub)
	AuthHeader = "auth_xpub"

	// AuthSignature is the given signature (body + timestamp)
	AuthSignature = "auth_signature"

	// AuthHeaderHash hash of the body coming from the request
	AuthHeaderHash = "auth_hash"

	// AuthHeaderNonce random nonce for the request
	AuthHeaderNonce = "auth_nonce"

	// AuthHeaderTime the time of the request, only valid for 30 seconds
	AuthHeaderTime = "auth_time"

	// AuthSignatureTTL is the max TTL for a signature to be valid
	AuthSignatureTTL = 20 * time.Second
)

// AuthPayload is the authentication payload for checking or creating a signature
type AuthPayload struct {
	AuthHash     string `json:"auth_hash"`
	AuthNonce    string `json:"auth_nonce"`
	AuthTime     int64  `json:"auth_time"`
	BodyContents string `json:"body_contents"`
	Signature    string `json:"signature"`
	xPub         string
}

// paramRequestKey for context key
type paramRequestKey string

const (
	xPubKey     paramRequestKey = "xpub"
	xPubHashKey paramRequestKey = "xpub_hash"
)

// createBodyHash will create the hash of the body, removing any carriage returns
func createBodyHash(bodyContents string) string {
	return utils.Hash(strings.TrimSuffix(bodyContents, "\n"))
}

// createSignature will create a signature for the given key & body contents
func createSignature(xPriv *bip32.ExtendedKey,
	bodyString string) (payload *AuthPayload, err error) {

	// No key?
	if xPriv == nil {
		err = ErrMissingXPriv
		return
	}

	// Get the xPub
	payload = new(AuthPayload)
	payload.xPub, err = bitcoin.GetExtendedPublicKey(xPriv)
	if err != nil { // Should never error if key is correct
		return
	}

	// Create the auth header hash
	payload.AuthHash = utils.Hash(bodyString)

	// auth_nonce is a random unique string to seed the signing message
	// this can be checked server side to make sure the request is not being replayed
	if payload.AuthNonce, err = utils.RandomHex(32); err != nil {
		return // Should never error
	}

	// Derive the address for signing
	var key *bip32.ExtendedKey
	key, err = utils.DeriveChildKeyFromHex(xPriv, payload.AuthNonce)
	if err != nil {
		return
	}

	var privateKey *bec.PrivateKey
	if privateKey, err = bitcoin.GetPrivateKeyFromHDKey(key); err != nil {
		return // Should never error if key is correct
	}

	// auth_time is the current time and makes sure a request can not be sent after 30 secs
	payload.AuthTime = time.Now().UnixMilli()

	// Signature, using bitcoin signMessage
	hexKey := hex.EncodeToString(privateKey.Serialise())
	payload.Signature, err = bitcoin.SignMessage(
		hexKey, fmt.Sprintf("%s%s%s%d", payload.xPub, payload.AuthHash, payload.AuthNonce, payload.AuthTime), true,
	)
	return
}

// setOnRequest will set the value on the request with the given key
func setOnRequest(req *http.Request, keyName paramRequestKey, value interface{}) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), keyName, value))
}

// getFromRequest gets the stored value from the request if found
func getFromRequest(req *http.Request, key paramRequestKey) (v string, ok bool) {
	v, ok = req.Context().Value(key).(string)
	return
}
