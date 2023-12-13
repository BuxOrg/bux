package chainstate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MerkleRootConfirmationState represents the state of each Merkle Root verification
// process and can be one of three values: Confirmed, Invalid and UnableToVerify.
type MerkleRootConfirmationState string

const (
	// Confirmed state occurs when Merkle Root is found in the longest chain.
	Confirmed MerkleRootConfirmationState = "CONFIRMED"
	// Invalid state occurs when Merkle Root is not found in the longest chain.
	Invalid MerkleRootConfirmationState = "INVALID"
	// UnableToVerify state occurs when Pulse is behind in synchronization with the longest chain.
	UnableToVerify MerkleRootConfirmationState = "UNABLE_TO_VERIFY"
)

// MerkleRootConfirmationRequestItem is a request type for verification
// of Merkle Roots inclusion in the longest chain.
type MerkleRootConfirmationRequestItem struct {
	MerkleRoot  string `json:"merkleRoot"`
	BlockHeight uint64 `json:"blockHeight"`
}

// MerkleRootConfirmation is a confirmation
// of merkle roots inclusion in the longest chain.
type MerkleRootConfirmation struct {
	Hash         string                      `json:"blockHash"`
	BlockHeight  uint64                      `json:"blockHeight"`
	MerkleRoot   string                      `json:"merkleRoot"`
	Confirmation MerkleRootConfirmationState `json:"confirmation"`
}

// MerkleRootsConfirmationsResponse is an API response for confirming
// merkle roots inclusion in the longest chain.
type MerkleRootsConfirmationsResponse struct {
	ConfirmationState MerkleRootConfirmationState `json:"confirmationState"`
	Confirmations     []MerkleRootConfirmation    `json:"confirmations"`
}

type pulseClientProvider struct {
	url       string
	authToken string
}

// verifyMerkleProof using Pulse
func (p pulseClientProvider) verifyMerkleRoots(
	ctx context.Context,
	c *Client, merkleRoots []MerkleRootConfirmationRequestItem,
) (*MerkleRootsConfirmationsResponse, error) {
	jsonData, err := json.Marshal(merkleRoots)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", p.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	if p.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.authToken)
	}
	res, err := client.Do(req)
	if err != nil {
		c.options.logger.Error().Msgf("Error during creating connection to pulse client: %s", err.Error())
		return nil, err
	}
	defer res.Body.Close() //nolint: all // Close the body

	// Parse response body.
	var merkleRootsRes MerkleRootsConfirmationsResponse
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error during reading response body: %s", err.Error())
	}

	err = json.Unmarshal(bodyBytes, &merkleRootsRes)
	if err != nil {
		return nil, fmt.Errorf("error during unmarshaling response body: %s", err.Error())
	}

	return &merkleRootsRes, nil
}
