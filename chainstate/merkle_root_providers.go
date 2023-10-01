package chainstate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/libsv/bitcoin-hc/transports/http/endpoints/api/merkleroots"
)

// Pulse provider
type pulseClientProvider struct {
	url       string
	authToken string
}

func (provider pulseClientProvider) getName() string {
	return ProviderPulse
}

// verifyMerkleProof using Pulse
func (provider pulseClientProvider) verifyMerkleRoots(ctx context.Context, merkleProofs []string) error {
	return verifyWithPulse(ctx, provider, merkleProofs)
}

func verifyWithPulse(ctx context.Context, provider pulseClientProvider, merkleProofs []string) error {
	jsonData, err := json.Marshal(merkleProofs)

	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", provider.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", provider.authToken)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint: all // Close the body

	// Parse response body.
	var merkleRootsRes merkleroots.MerkleRootsConfirmationsResponse
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error during reading response body: %s", err.Error())
	}

	err = json.Unmarshal(bodyBytes, &merkleRootsRes)
	if err != nil {
		return fmt.Errorf("error during unmarshaling response body: %s", err.Error())
	}

	if merkleRootsRes.AllConfirmed {
		return nil
	}
	return errors.New("not all merkle roots confirmed")
}
