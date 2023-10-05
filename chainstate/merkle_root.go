package chainstate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/merkleroots"
)

// VerifyMerkleRoots will try to verify merkle roots with all available providers
func (c *Client) VerifyMerkleRoots(ctx context.Context, merkleRoots []string) (*merkleroots.MerkleRootsConfirmationsResponse, error) {
	merkleRootsRes, err := verifyMerkleRootsWithPulse(ctx, c, merkleRoots)

	if err != nil {
		debugLog(c, "", fmt.Sprintf("verify merkle root error: %s from Pulse", err))
		return nil, err
	}
	debugLog(c, "", "successful verification of merkle proofs by Pulse")
	return merkleRootsRes, nil
}

func verifyMerkleRootsWithPulse(ctx context.Context, c *Client, merkleProofs []string) (*merkleroots.MerkleRootsConfirmationsResponse, error) {
	jsonData, err := json.Marshal(merkleProofs)

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", c.options.config.pulseClient.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.options.config.pulseClient.authToken)
	res, err := client.Do(req)
	if err != nil {
		c.options.logger.Error(context.Background(), "Error during creating connection to pulse client: %s", err.Error())
		return nil, err
	}
	defer res.Body.Close() //nolint: all // Close the body

	// Parse response body.
	var merkleRootsRes merkleroots.MerkleRootsConfirmationsResponse
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error during reading response body: %s", err.Error())
	}

	err = json.Unmarshal(bodyBytes, &merkleRootsRes)
	if err != nil {
		return nil, fmt.Errorf("error during unmarshaling response body: %s", err.Error())
	}

	if !merkleRootsRes.AllConfirmed {
		c.options.logger.Warn(context.Background(), "Warn: Not all merkle roots confirmed")
		return &merkleRootsRes, nil
	}
	return &merkleRootsRes, nil
}
