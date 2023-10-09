package chainstate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MerkleRootConfirmation is a confirmation
// of merkle roots inclusion in the longest chain.
type MerkleRootConfirmation struct {
	Hash       string `json:"blockhash"`
	MerkleRoot string `json:"merkleRoot"`
	Confirmed  bool   `json:"confirmed"`
}

// MerkleRootsConfirmationsResponse is an API response for confirming
// merkle roots inclusion in the longest chain.
type MerkleRootsConfirmationsResponse struct {
	AllConfirmed  bool                     `json:"allConfirmed"`
	Confirmations []MerkleRootConfirmation `json:"confirmations"`
}

type pulseClientProvider struct {
	url       string
	authToken string
}

// verifyMerkleProof using Pulse
func (p pulseClientProvider) verifyMerkleRoots(ctx context.Context, c *Client, merkleProofs []string) (*MerkleRootsConfirmationsResponse, error) {
	jsonData, err := json.Marshal(merkleProofs)

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
		c.options.logger.Error(context.Background(), "Error during creating connection to pulse client: %s", err.Error())
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
